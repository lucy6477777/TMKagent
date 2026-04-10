package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/lucyliuu/mini-tmk-agent/config"
	"github.com/lucyliuu/mini-tmk-agent/internal/asr"
	"github.com/lucyliuu/mini-tmk-agent/internal/audio"
	"github.com/lucyliuu/mini-tmk-agent/internal/pipeline"
	"github.com/lucyliuu/mini-tmk-agent/internal/translate"
)

var (
	apiKeyFlag  string
	baseURLFlag string
)

func main() {
	root := &cobra.Command{
		Use:   "mini-tmk-agent",
		Short: "A simultaneous interpretation CLI agent by Timekettle",
		Long: `mini-tmk-agent provides real-time and file-based speech translation.

Supported languages: zh (Chinese), en (English), es (Spanish), ja (Japanese)

Prerequisites:
  macOS: brew install portaudio
  Linux: sudo apt install portaudio19-dev libportaudio2

Environment:
  OPENAI_API_KEY  (required) Your OpenAI API key`,
	}

	root.PersistentFlags().StringVar(&apiKeyFlag, "api-key", "", "OpenAI API key (overrides OPENAI_API_KEY env var)")
	root.PersistentFlags().StringVar(&baseURLFlag, "base-url", "", "OpenAI API base URL (overrides OPENAI_BASE_URL env var)")

	root.AddCommand(newStreamCmd())
	root.AddCommand(newTranscriptCmd())

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

func loadConfig() *config.Config {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
	cfg.Override(apiKeyFlag, baseURLFlag)
	return cfg
}

func newStreamCmd() *cobra.Command {
	var sourceLang, targetLang string

	cmd := &cobra.Command{
		Use:   "stream",
		Short: "Real-time microphone translation to terminal",
		Example: `  mini-tmk-agent stream --source-lang zh --target-lang en
  mini-tmk-agent stream --source-lang en --target-lang ja`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := loadConfig()

			ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
			defer stop()

			asrClient := asr.NewWhisperClient(cfg.APIKey, cfg.BaseURL)
			translateClient := translate.NewOpenAIClient(cfg.APIKey, cfg.BaseURL)

			return pipeline.RunStream(ctx, pipeline.StreamConfig{
				SourceLang: sourceLang,
				TargetLang: targetLang,
				VADConfig:  audio.DefaultVADConfig(),
			}, asrClient, translateClient)
		},
	}

	cmd.Flags().StringVar(&sourceLang, "source-lang", "zh", "Source language (zh|en|es|ja)")
	cmd.Flags().StringVar(&targetLang, "target-lang", "en", "Target language (zh|en|es|ja)")
	return cmd
}

func newTranscriptCmd() *cobra.Command {
	var filePath, outputPath, sourceLang string

	cmd := &cobra.Command{
		Use:   "transcript",
		Short: "Transcribe an audio file to a text file",
		Example: `  mini-tmk-agent transcript --file speech.mp3 --output result.txt
  mini-tmk-agent transcript --file audio.pcm --output out.txt --source-lang zh`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := loadConfig()

			asrClient := asr.NewWhisperClient(cfg.APIKey, cfg.BaseURL)
			return pipeline.RunTranscript(context.Background(), pipeline.TranscriptConfig{
				FilePath:   filePath,
				OutputPath: outputPath,
				SourceLang: sourceLang,
			}, asrClient)
		},
	}

	cmd.Flags().StringVar(&filePath, "file", "", "Input audio file (.wav, .mp3, .pcm)")
	cmd.Flags().StringVar(&outputPath, "output", "", "Output text file path")
	cmd.Flags().StringVar(&sourceLang, "source-lang", "", "Source language code (default: auto-detect)")
	_ = cmd.MarkFlagRequired("file")
	_ = cmd.MarkFlagRequired("output")

	log.SetFlags(0) // cleaner log output without timestamp prefix
	return cmd
}
