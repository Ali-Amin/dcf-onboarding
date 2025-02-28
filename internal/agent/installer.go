package agent

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"clever.secure-onboard.com/internal/config"
	"clever.secure-onboard.com/pkg/contracts"
	"clever.secure-onboard.com/pkg/interfaces"
	"github.com/apenella/go-ansible/v2/pkg/execute"
	"github.com/apenella/go-ansible/v2/pkg/execute/configuration"
	results "github.com/apenella/go-ansible/v2/pkg/execute/result/json"
	"github.com/apenella/go-ansible/v2/pkg/execute/result/transformer"
	"github.com/apenella/go-ansible/v2/pkg/playbook"
)

func RemoteInstall(cfg config.DaemonInfo, hosts []string, logger interfaces.Logger) error {
	inventory := strings.Builder{}
	for _, host := range hosts {
		inventory.WriteString(host)
		inventory.WriteString(",")
	}

	logger.Write(
		slog.LevelInfo,
		fmt.Sprintf("Installing DCF agents on hosts: %s", inventory.String()),
	)

	ansiblePlaybookOptions := &playbook.AnsiblePlaybookOptions{
		User:         "dcf",
		BecomeUser:   "root",
		BecomeMethod: "sudo",
		Inventory:    inventory.String(),
		ExtraVars: map[string]interface{}{
			"ansible_sudo_pass":                    "password_here",
			"ansible_ssh_pass":                     "password_here",
			"host_key_checking":                    false,
			string(contracts.AgentPath):            cfg.BinaryPath,
			string(contracts.AgentSystemdUnitPath): cfg.SystemdUnitPath,
			string(contracts.OnboarderURL):         cfg.OnboardingURL,
			string(contracts.CFGPath):              cfg.ConfigPath,
			string(contracts.PrivKeyPath):          cfg.PrivKeyPath,
			string(contracts.HederaPrivKeyPath):    cfg.HederaPrivKeyPath,
		},
	}

	playbookCMD := playbook.NewAnsiblePlaybookCmd(
		playbook.WithPlaybooks(cfg.PlaybookPath),
		playbook.WithPlaybookOptions(ansiblePlaybookOptions),
	)

	cmd, _ := playbookCMD.Command()
	logger.Write(slog.LevelInfo, fmt.Sprintf("Running command on hosts: %s", cmd))

	buff := new(bytes.Buffer)

	exec := configuration.NewAnsibleWithConfigurationSettingsExecute(
		execute.NewDefaultExecute(
			execute.WithCmd(playbookCMD),
			execute.WithTransformers(
				transformer.LogFormat(transformer.DefaultLogFormatLayout, transformer.Now),
			),
		),
		configuration.WithAnsiblePipelining(),
	)

	err := exec.Execute(context.Background())
	if err != nil {
		fmt.Println(err)
		return err
	}

	res, err := results.ParseJSONResultsStream(io.Reader(buff))
	if err != nil {
		panic(err)
	}

	logger.Write(slog.LevelDebug, res.String())
	return nil
}
