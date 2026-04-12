package audio

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
)

func TestPCMToWAV_HeaderBytes(t *testing.T) {
	pcm := make([]byte, 3200) // 100ms of silence at 16kHz 16-bit mono
	wav := PCMToWAV(pcm, 16000, 1, 16)

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
	data, filename, err := ReadAudioFile("../../testdata/hello_zh.wav")
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
	os.WriteFile(tmp, make([]byte, 3200), 0644)

	data, filename, err := ReadAudioFile(tmp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if filename != "audio.wav" {
		t.Errorf("PCM should be returned with filename audio.wav, got %q", filename)
	}
	if len(data) != 3200+44 {
		t.Errorf("expected %d bytes (PCM+header), got %d", 3200+44, len(data))
	}
}

func TestReadAudioFile_MP3(t *testing.T) {
	tmp := filepath.Join(t.TempDir(), "test.mp3")
	os.WriteFile(tmp, []byte("fake-mp3-data"), 0644)

	data, filename, err := ReadAudioFile(tmp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if filename != "audio.mp3" {
		t.Errorf("expected filename audio.mp3, got %q", filename)
	}
	if string(data) != "fake-mp3-data" {
		t.Errorf("MP3 data should be returned as-is")
	}
}

func TestReadAudioFile_M4A(t *testing.T) {
	tmp := filepath.Join(t.TempDir(), "test.m4a")
	os.WriteFile(tmp, []byte("fake-m4a-data"), 0644)

	data, filename, err := ReadAudioFile(tmp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if filename != "audio.m4a" {
		t.Errorf("expected filename audio.m4a, got %q", filename)
	}
	if string(data) != "fake-m4a-data" {
		t.Errorf("M4A data should be returned as-is")
	}
}

func TestReadAudioFile_UnsupportedFormat(t *testing.T) {
	tmp := filepath.Join(t.TempDir(), "test.ogg")
	os.WriteFile(tmp, []byte("fake"), 0644)

	_, _, err := ReadAudioFile(tmp)
	if err == nil {
		t.Error("expected error for unsupported format")
	}
}

func TestReadAudioFile_NonexistentFile(t *testing.T) {
	_, _, err := ReadAudioFile("/nonexistent/file.wav")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestInt16ToBytes_LittleEndian(t *testing.T) {
	samples := []int16{0x0102, 0x0304}
	b := Int16ToBytes(samples)
	if len(b) != 4 {
		t.Fatalf("expected 4 bytes, got %d", len(b))
	}
	// 0x0102 little-endian → [0x02, 0x01]
	if b[0] != 0x02 || b[1] != 0x01 {
		t.Errorf("little-endian encoding wrong: got %x %x", b[0], b[1])
	}
}

func TestInt16ToBytes_Empty(t *testing.T) {
	b := Int16ToBytes(nil)
	if len(b) != 0 {
		t.Errorf("expected empty, got %d bytes", len(b))
	}
}
