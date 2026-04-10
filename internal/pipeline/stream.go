package pipeline

import (
	"context"
	"fmt"
	"log"

	"github.com/lucyliuu/mini-tmk-agent/internal/asr"
	"github.com/lucyliuu/mini-tmk-agent/internal/audio"
	"github.com/lucyliuu/mini-tmk-agent/internal/display"
	"github.com/lucyliuu/mini-tmk-agent/internal/translate"
)

// StreamConfig holds parameters for the stream command.
type StreamConfig struct {
	SourceLang string
	TargetLang string
	VADConfig  audio.VADConfig
}

// RunStream starts the real-time microphone→ASR→translate→display pipeline.
// Blocks until ctx is cancelled (e.g. by Ctrl+C).
func RunStream(ctx context.Context, cfg StreamConfig, asrClient asr.Client, translateClient translate.Client) error {
	capturer, frameCh, err := audio.NewCapturer()
	if err != nil {
		return err
	}
	if err := capturer.Start(); err != nil {
		return fmt.Errorf("starting audio capture: %w", err)
	}
	defer capturer.Stop()

	fmt.Println("Listening... Press Ctrl+C to stop.\n")
	runStreamFromChannel(ctx, cfg, frameCh, asrClient, translateClient)
	return nil
}

// runStreamFromChannel runs the VAD→ASR→translate→display pipeline on the given frame channel.
// Extracted to allow testing without a real microphone.
func runStreamFromChannel(
	ctx context.Context,
	cfg StreamConfig,
	frameCh <-chan []int16,
	asrClient asr.Client,
	translateClient translate.Client,
) {
	audioCh := make(chan []int16, 8)
	asrCh := make(chan string, 8)
	translateCh := make(chan display.Pair, 8)

	vad := audio.NewVAD(cfg.VADConfig)

	// g1: VAD — accumulate frames, emit speech chunks
	go func() {
		defer close(audioCh)
		for {
			select {
			case <-ctx.Done():
				return
			case frame, ok := <-frameCh:
				if !ok {
					return
				}
				if chunk := vad.Feed(frame); chunk != nil {
					select {
					case audioCh <- chunk:
					case <-ctx.Done():
						return
					}
				}
			}
		}
	}()

	// g2: ASR — transcribe audio chunks
	go func() {
		defer close(asrCh)
		for chunk := range audioCh {
			pcmBytes := audio.Int16ToBytes(chunk)
			wavBytes := audio.PCMToWAV(pcmBytes, 16000, 1, 16)
			text, err := asrClient.Transcribe(ctx, wavBytes, "audio.wav", cfg.SourceLang)
			if err != nil {
				log.Printf("[WARN] ASR error (skipping chunk): %v", err)
				continue
			}
			if text == "" {
				continue
			}
			select {
			case asrCh <- text:
			case <-ctx.Done():
				return
			}
		}
	}()

	// g3: Translate — translate recognised text
	go func() {
		defer close(translateCh)
		for text := range asrCh {
			translated, err := translateClient.Translate(ctx, text, cfg.SourceLang, cfg.TargetLang)
			if err != nil {
				log.Printf("[WARN] translation error: %v", err)
				translated = "[翻译失败]"
			}
			select {
			case translateCh <- display.Pair{Source: text, Target: translated}:
			case <-ctx.Done():
				return
			}
		}
	}()

	// g4: Display — render to terminal (blocks until translateCh closed)
	for pair := range translateCh {
		display.Print(pair)
	}
}
