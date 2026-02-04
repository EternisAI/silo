package inference

import (
	"os"
	"strings"
	"testing"

	"github.com/eternisai/silo/internal/config"
	"github.com/eternisai/silo/pkg/logger"
)

func TestBuildDockerRunArgs(t *testing.T) {
	cfg := &config.Config{}
	log := logger.New(false)
	engine := New(cfg, log)

	args := engine.buildDockerRunArgs()
	cmd := "docker " + strings.Join(args, " ")

	// Expected command (with $HOME expanded)
	homeDir := os.Getenv("HOME")
	expected := `docker run -d --name glm_model --restart unless-stopped --gpus "device=0,1,2" --shm-size 64g --ipc=host --ulimit memlock=-1:-1 --ulimit nofile=1048576:1048576 -p 30000:30000 -e CUDA_VISIBLE_DEVICES=0,1,2 -e PYTORCH_ALLOC_CONF=expandable_segments:True -v /root/data/AWQ:/workspace/model -v ` + homeDir + `/.cache/huggingface:/root/.cache/huggingface lmsysorg/sglang:latest python3 -m sglang.launch_server --model-path /workspace/model --host 0.0.0.0 --port 30000 --dp-size 3 --tp-size 1 --max-running-requests 32 --max-total-tokens 262144 --context-length 131072 --mem-fraction-static 0.88 --chunked-prefill-size 8192 --schedule-policy fcfs --kv-cache-dtype fp8_e4m3 --attention-backend flashinfer --disable-radix-cache --reasoning-parser glm45 --tool-call-parser glm --trust-remote-code --log-level info`

	if cmd != expected {
		t.Errorf("Docker command mismatch.\n\nGot:\n%s\n\nExpected:\n%s", cmd, expected)
	}
}

func TestBuildDockerRunArgs_ContainerName(t *testing.T) {
	cfg := &config.Config{}
	log := logger.New(false)
	engine := New(cfg, log)

	args := engine.buildDockerRunArgs()

	// Find --name argument
	for i, arg := range args {
		if arg == "--name" && i+1 < len(args) {
			if args[i+1] != "glm_model" {
				t.Errorf("Container name should be 'glm_model', got '%s'", args[i+1])
			}
			return
		}
	}
	t.Error("--name argument not found")
}

func TestBuildDockerRunArgs_SGLangFlags(t *testing.T) {
	cfg := &config.Config{}
	log := logger.New(false)
	engine := New(cfg, log)

	args := engine.buildDockerRunArgs()
	cmd := strings.Join(args, " ")

	requiredFlags := []string{
		"--dp-size 3",
		"--tp-size 1",
		"--max-running-requests 32",
		"--max-total-tokens 262144",
		"--context-length 131072",
		"--mem-fraction-static 0.88",
		"--chunked-prefill-size 8192",
		"--schedule-policy fcfs",
		"--kv-cache-dtype fp8_e4m3",
		"--attention-backend flashinfer",
		"--disable-radix-cache",
		"--reasoning-parser glm45",
		"--tool-call-parser glm",
		"--trust-remote-code",
		"--log-level info",
	}

	for _, flag := range requiredFlags {
		if !strings.Contains(cmd, flag) {
			t.Errorf("Missing required flag: %s", flag)
		}
	}
}
