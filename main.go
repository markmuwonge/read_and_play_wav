package main

import (
	"log"
	"os"
	custom_error "read_and_play_wav/error"

	"github.com/hajimehoshi/oto"
	"github.com/thoas/go-funk"
	"github.com/tidwall/gjson"
	"github.com/youpy/go-wav"
)

func main() {
	config_bytes, err := os.ReadFile("config.json")
	custom_error.Fatal(err)
	wav_file_path := gjson.Get(string(config_bytes), "wav_file_path").String()
	// playWav(wav_file_path)
	streamWav(wav_file_path)
}

func isPCM(wav_file *os.File) (*wav.Reader, bool) {
	wav_file_reader := wav.NewReader(wav_file)
	wav_file_format, err := wav_file_reader.Format()
	custom_error.Fatal(err)
	if wav_file_format.AudioFormat != wav.AudioFormatPCM {
		log.Println("WAV Format is not PCM")
		return nil, false
	}

	return wav_file_reader, true
}

func playWav(wav_file_path string) {
	wav_file, err := os.Open(wav_file_path)
	defer wav_file.Close()
	custom_error.Fatal(err)

	wav_file_reader, pcm := isPCM(wav_file)
	if !pcm {
		return
	}

	_, err = wav_file_reader.ReadSamples()
	custom_error.Fatal(err)
	wav_file_data_bytes := make([]byte, wav_file_reader.WavData.Size)
	wav_file.ReadAt(wav_file_data_bytes, 44)

	wav_file_format, err := wav_file_reader.Format()
	custom_error.Fatal(err)

	wav_file_data_byte_chunks := funk.Chunk(wav_file_data_bytes, int(wav_file_format.SampleRate)).([][]byte)

	oto_ctx, err := oto.NewContext(int(wav_file_format.SampleRate), int(wav_file_format.NumChannels), int(wav_file_format.BitsPerSample/8), int(wav_file_format.SampleRate))
	custom_error.Fatal(err)
	player := oto_ctx.NewPlayer()
	defer player.Close()

	for _, v := range wav_file_data_byte_chunks {
		player.Write(v)
	}
}

func streamWav(wav_file_path string) {
	wav_file, err := os.Open(wav_file_path)
	defer wav_file.Close()
	custom_error.Fatal(err)

	wav_file_reader, pcm := isPCM(wav_file)
	if !pcm {
		return
	}

	_, err = wav_file_reader.ReadSamples(1)
	custom_error.Fatal(err)
	wav_file_data_size := wav_file_reader.WavData.Size

	wav_file_data_indexes := func() []int {
		ret := make([]int, 0)
		for i := 44; i <= int(44+(wav_file_data_size-1)); i++ {
			ret = append(ret, i)
		}
		return ret
	}()

	wav_file_format, err := wav_file_reader.Format()
	custom_error.Fatal(err)
	wav_file_sample_rate := wav_file_format.SampleRate

	wav_file_data_index_chunks := funk.Chunk(wav_file_data_indexes, int(wav_file_sample_rate)).([][]int)
	if len(wav_file_data_index_chunks[len(wav_file_data_index_chunks)-1]) != int(wav_file_sample_rate) {
		wav_file_data_index_chunks = funk.Initial(wav_file_data_index_chunks).([][]int)
	}

	oto_ctx, err := oto.NewContext(int(wav_file_format.SampleRate), int(wav_file_format.NumChannels), int(wav_file_format.BitsPerSample/8), int(wav_file_format.SampleRate))
	custom_error.Fatal(err)
	player := oto_ctx.NewPlayer()
	defer player.Close()

	buffer := make([]byte, wav_file_sample_rate)
	for _, wav_file_data_index_chunk := range wav_file_data_index_chunks {
		index := wav_file_data_index_chunk[0]
		_, err := wav_file.ReadAt(buffer, int64(index))
		custom_error.Fatal(err)
		player.Write(buffer)
	}
}
