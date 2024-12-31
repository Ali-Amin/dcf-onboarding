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
	"github.com/apenella/go-ansible/v2/pkg/execute/measure"
	results "github.com/apenella/go-ansible/v2/pkg/execute/result/json"
	"github.com/apenella/go-ansible/v2/pkg/execute/stdoutcallback"
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
		User:         "ubunu",
		BecomeUser:   "root",
		BecomeMethod: "sudo",
		Inventory:    inventory.String(),
		ExtraVars: map[string]interface{}{
			"ansible_sudo_pass":                    "ubuntu",
			string(contracts.AgentPath):            cfg.BinaryPath,
			string(contracts.AgentSystemdUnitPath): cfg.SystemdUnitPath,
			string(contracts.OnboarderURL):         cfg.OnboardingURL,
			string(contracts.CFGPath):              cfg.ConfigPath,
			string(contracts.PrivKeyPath):          cfg.PrivKeyPath,
		},
	}

	playbookCMD := playbook.NewAnsiblePlaybookCmd(
		playbook.WithPlaybooks(cfg.PlaybookPath),
		playbook.WithPlaybookOptions(ansiblePlaybookOptions),
	)

	cmd, _ := playbookCMD.Command()
	logger.Write(slog.LevelInfo, fmt.Sprintf("Running command on hosts: %s", cmd))

	buff := new(bytes.Buffer)
	exec := measure.NewExecutorTimeMeasurement(
		stdoutcallback.NewJSONStdoutCallbackExecute(
			execute.NewDefaultExecute(
				execute.WithCmd(playbookCMD),
				execute.WithErrorEnrich(playbook.NewAnsiblePlaybookErrorEnrich()),
				execute.WithWrite(io.Writer(buff)),
			),
		),
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

	fmt.Println(res.String())
	fmt.Println("Duration: ", exec.Duration().String())
	return nil
}
