package unit_test

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"

	"github.com/lucyliuu/mini-tmk-agent/internal/audio"
)

func TestPCMToWAV_HeaderBytes(t *testing.T) {
	pcm := make([]byte, 3200) // 100ms of silence at 16kHz 16-bit mono
	wav := audio.PCMToWAV(pcm, 16000, 1, 16)

	if len(wav) != 44+len(pcm) {
		t.Fatalf("expected %d bytes, got %d", 44+len(pcm), len(wav))
	}
	if string(wav[0:4]) != "RIFF" {
		t.Errorf("expected RIFF marker, got %q", wav[0:4])
	}
	if string(wav[8:12]) != "WAVE" {
		t.Errorf("expected WAVE marker, got %q", wav[8:12])
	}
	if string(wav[12:16]) != "fmt " {
		t.Errorf("expected fmt  marker, got %q", wav[12:16])
	}
	sr := binary.LittleEndian.Uint32(wav[24:28])
	if sr != 16000 {
		t.Errorf("expected sample rate 16000, got %d", sr)
	}
	if string(wav[36:40]) != "data" {
		t.Errorf("expected data marker, got %q", wav[36:40])
	}
	dataSize := binary.LittleEndian.Uint32(wav[40:44])
	if dataSize != uint32(len(pcm)) {
		t.Errorf("expected data size %d, got %d", len(pcm), dataSize)
	}
}

func TestReadAudioFile_WAV(t *testing.T) {
	data, filename, err := audio.ReadAudioFile("../../testdata/hello_zh.wav")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if filename != "audio.wav" {
		t.Errorf("expected filename audio.wav, got %q", filename)
	}
	if len(data) == 0 {
		t.Error("expected non-empty data")
	}
}

func TestReadAudioFile_PCM(t *testing.T) {
	tmp := filepath.Join(t.TempDir(), "test.pcm")
	pcm := make([]byte, 3200)
	os.WriteFile(tmp, pcm, 0644)

	data, filename, err := audio.ReadAudioFile(tmp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if filename != "audio.wav" {
		t.Errorf("PCM should be returned with filename audio.wav, got %q", filename)
	}
	if len(data) != len(pcm)+44 {
		t.Errorf("expected %d bytes (PCM+header), got %d", len(pcm)+44, len(data))
	}
}

func TestReadAudioFile_UnsupportedFormat(t *testing.T) {
	tmp := filepath.Join(t.TempDir(), "test.ogg")
	os.WriteFile(tmp, []byte("fake"), 0644)

	_, _, err := audio.ReadAudioFile(tmp)
	if err == nil {
		t.Error("expected error for unsupported format")
	}
}

func TestInt16ToBytes_LittleEndian(t *testing.T) {
	samples := []int16{0x0102, 0x0304}
	b := audio.Int16ToBytes(samples)
	if len(b) != 4 {
		t.Fatalf("expected 4 bytes, got %d", len(b))
	}
	// 0x0102 little-endian → [0x02, 0x01]
	if b[0] != 0x02 || b[1] != 0x01 {
		t.Errorf("little-endian encoding wrong: got %x %x", b[0], b[1])
	}
}
