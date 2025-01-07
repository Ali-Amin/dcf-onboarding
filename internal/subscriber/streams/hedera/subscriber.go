package hedera

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"sync"

	"github.com/hashgraph/hedera-sdk-go/v2"
	"github.com/project-alvarium/alvarium-sdk-go/pkg/config"
	"github.com/project-alvarium/alvarium-sdk-go/pkg/contracts"
	"github.com/project-alvarium/alvarium-sdk-go/pkg/interfaces"
	"github.com/project-alvarium/alvarium-sdk-go/pkg/message"
)

type hederaSubscriber struct {
	cfg    config.HederaConfig
	pub    chan message.SubscribeWrapper
	client *hedera.Client
	logger interfaces.Logger
}

func NewHederaSubscriber(
	cfg config.HederaConfig,
	pub chan message.SubscribeWrapper,
	logger interfaces.Logger,
) (*hederaSubscriber, error) {
	// client init
	var client *hedera.Client
	switch cfg.NetType {
	case contracts.Local:
		node := make(map[string]hedera.AccountID, 1)
		consensus := cfg.Consensus.Address()
		node[consensus] = hedera.AccountID{Account: 3}

		mirror := cfg.Mirror.Address()
		mirrorNode := []string{mirror}

		client = hedera.ClientForNetwork(node)
		client.SetMirrorNetwork(mirrorNode)

	case contracts.Mainnet:
		client = hedera.ClientForMainnet()
	case contracts.Testnet:
		client = hedera.ClientForTestnet()
	case contracts.Previewnet:
		client = hedera.ClientForPreviewnet()
	}

	// Set client operator
	accountId, err := hedera.AccountIDFromString(cfg.AccountId)
	if err != nil {
		return nil, err
	}

	privateKey, err := readPrivateKey(cfg)
	if err != nil {
		return nil, err
	}

	client.SetOperator(accountId, privateKey)

	// default values
	maxTxFeeInHbar := hedera.HbarFrom(cfg.DefaultMaxTxFee, hedera.HbarUnits.Hbar)
	maxQueryPaymentFee := hedera.HbarFrom(cfg.DefaultMaxQueryPayment, hedera.HbarUnits.Hbar)

	err = client.SetDefaultMaxTransactionFee(maxTxFeeInHbar)
	if err != nil {
		return nil, err
	}

	err = client.SetDefaultMaxQueryPayment(maxQueryPaymentFee)
	if err != nil {
		return nil, err
	}

	return &hederaSubscriber{
		client: client,
		cfg:    cfg,
		pub:    pub,
		logger: logger,
	}, nil
}

func (h *hederaSubscriber) Subscribe(ctx context.Context, wg *sync.WaitGroup) bool {
	for _, topic := range h.cfg.Topics {
		h.logger.Write(
			slog.LevelDebug,
			fmt.Sprintf("Subscribing to topic %s", topic),
		)

		topicID, err := hedera.TopicIDFromString(topic)
		if err != nil {
			h.logger.Error("Failed to subscribe to topic: " + err.Error())
			continue
		}

		_, err = hedera.NewTopicMessageQuery().
			SetTopicID(topicID).
			Subscribe(h.client, func(msg hedera.TopicMessage) {
				h.logger.Write(
					slog.LevelDebug,
					fmt.Sprintf(
						"Found message %s on topic %s: %s",
						msg.TransactionID,
						topic,
						string(msg.Contents),
					),
				)

				var wrapper message.SubscribeWrapper
				annotationListType := fmt.Sprintf("%T", []contracts.AnnotationList{})
				err := json.Unmarshal(msg.Contents, &wrapper)
				if err != nil || wrapper.MessageType != annotationListType {
					h.logger.Write(
						slog.LevelWarn,
						"skipping message, not an annotation list",
					)
				} else {
					h.pub <- wrapper
				}
			})
		if err != nil {
			h.logger.Error(fmt.Sprintf("Failed to subscribe to topic '%s': %s", topic, err.Error()))
		}

	}

	wg.Add(1)
	go func() { // Graceful shutdown
		defer wg.Done()

		<-ctx.Done()
		close(h.pub)
		h.client.Close()
		h.logger.Write(slog.LevelInfo, "shutdown received")
	}()

	return true
}

func (h *hederaSubscriber) Close() {
	h.client.Close()
}

func readPrivateKey(cfg config.HederaConfig) (hedera.PrivateKey, error) {
	b, err := os.ReadFile(cfg.PrivateKeyPath)
	if err != nil {
		return hedera.PrivateKey{}, err
	}

	// It was reported by multiple parties that a `\n` character is
	// occasionally loaded into the private key byte array, and
	// other times it was not when the file was created using
	// different methods or saved by different editors
	//
	// This part will remove the newline character only if it exists
	privateKeyDER := string(b)
	if privateKeyDER[len(privateKeyDER)-1] == '\n' {
		privateKeyDER = privateKeyDER[:len(privateKeyDER)-1]
	}

	privateKey, err := hedera.PrivateKeyFromStringDer(privateKeyDER)
	if err != nil {
		return hedera.PrivateKey{}, err
	}

	return privateKey, nil
}
