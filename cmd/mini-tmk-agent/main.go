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
	"github.com/lucyliuu/mini-tmk-agent/internal/pipeline"
	"github.com/lucyliuu/mini-tmk-agent/internal/translate"
	"github.com/lucyliuu/mini-tmk-agent/internal/tts"
)

var (
	apiKeyFlag         string
	baseURLFlag        string
	deepgramAPIKeyFlag string
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
  OPENAI_API_KEY   (required) Your OpenAI API key
  DEEPGRAM_API_KEY (required for stream) Deepgram API key for streaming ASR`,
	}

	root.PersistentFlags().StringVar(&apiKeyFlag, "api-key", "", "OpenAI API key (overrides OPENAI_API_KEY env var)")
	root.PersistentFlags().StringVar(&baseURLFlag, "base-url", "", "OpenAI API base URL (overrides OPENAI_BASE_URL env var)")
	root.PersistentFlags().StringVar(&deepgramAPIKeyFlag, "deepgram-api-key", "", "Deepgram API key (overrides DEEPGRAM_API_KEY env var)")

	root.AddCommand(newStreamCmd())
	root.AddCommand(newTranscriptCmd())
	root.AddCommand(newWebCmd())

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
	cfg.Override(apiKeyFlag, baseURLFlag, deepgramAPIKeyFlag)
	return cfg
}

func newStreamCmd() *cobra.Command {
	var (
		sourceLang     string
		targetLang     string
		enableTTS      bool
		ttsVoice       string
		ttsSpeakerMode bool
	)

	cmd := &cobra.Command{
		Use:   "stream",
		Short: "Real-time microphone translation to terminal",
		Example: `  # Streaming translation (requires DEEPGRAM_API_KEY)
  mini-tmk-agent stream --source-lang zh --target-lang en

  # With TTS output (requires headphones for full-duplex)
  mini-tmk-agent stream --source-lang zh --target-lang en --tts

  # TTS with speakers (half-duplex: mic pauses during playback)
  mini-tmk-agent stream --source-lang zh --target-lang en --tts --tts-speaker-mode`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := loadConfig()

			if cfg.DeepgramAPIKey == "" {
				return fmt.Errorf(
					"DEEPGRAM_API_KEY is required for stream mode\n" +
						"  Sign up at https://console.deepgram.com ($200 free credits)\n" +
						"  Run: export DEEPGRAM_API_KEY=dg-...\n" +
						"  Or:  mini-tmk-agent --deepgram-api-key dg-...",
				)
			}

			ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
			defer stop()

			translateClient := translate.NewOpenAIClient(cfg.APIKey, cfg.BaseURL)

			streamCfg := pipeline.StreamConfig{
				SourceLang: sourceLang,
				TargetLang: targetLang,
				StreamASR:  asr.NewDeepgramStreamClient(cfg.DeepgramAPIKey),
			}

			if enableTTS {
				streamCfg.EnableTTS = true
				streamCfg.TTSSpeakerMode = ttsSpeakerMode
				streamCfg.TTSClient = tts.NewOpenAIClient(cfg.APIKey, cfg.BaseURL, ttsVoice, "pcm")
				streamCfg.TTSPlayer = tts.NewPlayer(true)
				if ttsSpeakerMode {
					fmt.Println("TTS enabled (speaker mode: mic pauses during playback)")
				} else {
					fmt.Println("TTS enabled (use headphones for full-duplex)")
				}
			}

			return pipeline.RunStream(ctx, streamCfg, translateClient)
		},
	}

	cmd.Flags().StringVar(&sourceLang, "source-lang", "zh", "Source language (zh|en|es|ja)")
	cmd.Flags().StringVar(&targetLang, "target-lang", "en", "Target language (zh|en|es|ja)")
	cmd.Flags().BoolVar(&enableTTS, "tts", false, "Enable TTS audio output for translations")
	cmd.Flags().StringVar(&ttsVoice, "tts-voice", "alloy", "TTS voice (alloy|echo|fable|onyx|nova|shimmer)")
	cmd.Flags().BoolVar(&ttsSpeakerMode, "tts-speaker-mode", false, "Pause mic during TTS playback (use with speakers, not headphones)")
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

	log.SetFlags(0)
	return cmd
}
