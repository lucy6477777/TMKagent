package pipeline

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/lucyliuu/mini-tmk-agent/internal/asr"
	"github.com/lucyliuu/mini-tmk-agent/internal/audio"
	"github.com/lucyliuu/mini-tmk-agent/internal/display"
	"github.com/lucyliuu/mini-tmk-agent/internal/translate"
	"github.com/lucyliuu/mini-tmk-agent/internal/tts"
)

// StreamConfig holds parameters for the stream command.
type StreamConfig struct {
	SourceLang string
	TargetLang string

	// Streaming ASR (Deepgram).
	StreamASR asr.StreamClient

	// TTS options. EnableTTS activates text-to-speech output.
	EnableTTS      bool
	TTSSpeakerMode bool // when true, mic input pauses during TTS playback
	TTSClient      tts.Client
	TTSPlayer      *tts.Player
}

// RunStream starts the real-time microphone→ASR→translate→display pipeline.
// Blocks until ctx is cancelled (e.g. by Ctrl+C).
func RunStream(ctx context.Context, cfg StreamConfig, translateClient translate.Client) error {
	capturer, frameCh, err := audio.NewCapturer()
	if err != nil {
		return err
	}
	if err := capturer.Start(); err != nil {
		return fmt.Errorf("starting audio capture: %w", err)
	}
	defer capturer.Stop()

	ml, logFile, err := newMetricsLogger()
	if err != nil {
		log.Printf("[WARN] could not create metrics log: %v", err)
		ml = nil
	} else {
		fmt.Printf("Logging metrics to %s\n", logFile)
		defer ml.close()
	}

	fmt.Println("Listening... Press Ctrl+C to stop.")
	return runStreamingPipeline(ctx, cfg, frameCh, translateClient, ml)
}

func runStreamingPipeline(
	ctx context.Context,
	cfg StreamConfig,
	frameCh <-chan []int16,
	translateClient translate.Client,
	ml *metricsLogger,
) error {
	session, err := cfg.StreamASR.Connect(ctx, cfg.SourceLang)
	if err != nil {
		return fmt.Errorf("connecting to streaming ASR: %w", err)
	}
	defer session.Close()

	var ttsCh chan string
	if cfg.EnableTTS && cfg.TTSClient != nil && cfg.TTSPlayer != nil {
		ttsCh = make(chan string, 3)
		go runTTSWorker(ctx, cfg.TTSClient, cfg.TTSPlayer, ttsCh)
	}

	// G1: send audio frames to Deepgram
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case frame, ok := <-frameCh:
				if !ok {
					return
				}
				if cfg.TTSSpeakerMode && cfg.TTSPlayer != nil && cfg.TTSPlayer.IsPlaying() {
					continue
				}
				if err := session.Send(audio.Int16ToBytes(frame)); err != nil {
					log.Printf("[WARN] send audio: %v", err)
					return
				}
			}
		}
	}()

	// Main goroutine: read ASR results → translate → display → TTS
	w := display.NewWriter()
	for result := range session.Results() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if !result.IsFinal {
			w.PrintInterim(result.Text)
			continue
		}

		w.ClearInterim()

		t1 := time.Now()
		normalizedSrc, translated, err := translateClient.Translate(ctx, result.Text, cfg.SourceLang, cfg.TargetLang)
		if ml != nil {
			ml.logTranslate(
				time.Since(t1).Milliseconds(),
				int64(len([]rune(result.Text))),
				int64(len([]rune(translated))),
			)
		}
		if err != nil {
			log.Printf("[WARN] translation error: %v", err)
			translated = "[翻译失败]"
		}

		src := result.Text
		if normalizedSrc != "" {
			src = normalizedSrc
		}

		w.PrintFinal(display.Pair{Source: src, Target: translated})

		if ttsCh != nil && translated != "[翻译失败]" {
			select {
			case ttsCh <- translated:
			default:
			}
		}
	}
	return nil
}

// ---------- TTS worker ----------

func runTTSWorker(ctx context.Context, client tts.Client, player *tts.Player, ch <-chan string) {
	for {
		select {
		case <-ctx.Done():
			return
		case text, ok := <-ch:
			if !ok {
				return
			}
			stream, err := client.Speak(ctx, text)
			if err != nil {
				log.Printf("[WARN] TTS error: %v", err)
				continue
			}
			if err := player.PlayStream(stream); err != nil {
				log.Printf("[WARN] TTS playback error: %v", err)
			}
			stream.Close()
		}
	}
}
