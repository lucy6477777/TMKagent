package audio

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ReadAudioFile reads an audio file and returns bytes ready for the Whisper API.
// WAV and MP3 are returned as-is with their original extension as filename.
// PCM is wrapped with a 44-byte WAV header (assumes 16kHz, mono, 16-bit).
func ReadAudioFile(path string) ([]byte, string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, "", fmt.Errorf("reading %s: %w", path, err)
	}
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".wav":
		return data, "audio.wav", nil
	case ".mp3":
		return data, "audio.mp3", nil
	case ".pcm":
		return PCMToWAV(data, 16000, 1, 16), "audio.wav", nil
	default:
		return nil, "", fmt.Errorf("unsupported format %q; supported: .wav, .mp3, .pcm", ext)
	}
}

// PCMToWAV wraps raw PCM bytes in a RIFF/WAV header.
// Parameters: sampleRate (Hz), channels (1=mono), bitsPerSample (16 typical).
func PCMToWAV(pcm []byte, sampleRate, channels, bitsPerSample int) []byte {
	dataSize := len(pcm)
	buf := make([]byte, 44+dataSize)

	copy(buf[0:4], "RIFF")
	binary.LittleEndian.PutUint32(buf[4:8], uint32(36+dataSize))
	copy(buf[8:12], "WAVE")

	copy(buf[12:16], "fmt ")
	binary.LittleEndian.PutUint32(buf[16:20], 16)
	binary.LittleEndian.PutUint16(buf[20:22], 1) // PCM audio format
	binary.LittleEndian.PutUint16(buf[22:24], uint16(channels))
	binary.LittleEndian.PutUint32(buf[24:28], uint32(sampleRate))
	byteRate := sampleRate * channels * bitsPerSample / 8
	binary.LittleEndian.PutUint32(buf[28:32], uint32(byteRate))
	blockAlign := channels * bitsPerSample / 8
	binary.LittleEndian.PutUint16(buf[32:34], uint16(blockAlign))
	binary.LittleEndian.PutUint16(buf[34:36], uint16(bitsPerSample))

	copy(buf[36:40], "data")
	binary.LittleEndian.PutUint32(buf[40:44], uint32(dataSize))
	copy(buf[44:], pcm)
	return buf
}

// Int16ToBytes converts []int16 samples to []byte (little-endian).
func Int16ToBytes(samples []int16) []byte {
	buf := make([]byte, len(samples)*2)
	for i, s := range samples {
		binary.LittleEndian.PutUint16(buf[i*2:], uint16(s))
	}
	return buf
}
