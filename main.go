package main

import "fmt"
import "flag"
import "os"
import "path"
import "regexp"
import "log"

var (
	url             = flag.String("url", "", "the url to download using cobalt.\n")
	videoCodec      = flag.String("video-codec", "h264", "Video codec to be used. Applies only to youtube downloads. AV1: 8K/HDR, lower support | VP9: 4K/HDR, best quality | H264: 1080p, works everywhere\n")
	videoQuality    = flag.Int("video-quality", 1080, "Quality of the video, also applies only to youtube downloads. Ranges from 144p to 2160p (4k).\n")
	audioCodec      = flag.String("audio-codec", "best", "Audio format/codec to be used. \"best\" doesn't re-encodes the audio.\n")
	filenamePattern = flag.String("filename", "pretty", "File name pattern. Classic: youtube_yPYZpwSpKmA_1920x1080_h264.mp4 | audio: youtube_yPYZpwSpKmA_audio.mp3 // Basic: Video Title (1080p, h264).mp4 | audio: Audio Title - Audio Author.mp3 // Pretty: Video Title (1080p, h264, youtube).mp4 | audio: Audio Title - Audio Author (soundcloud).mp3 // Nerdy: Video Title (1080p, h264, youtube, yPYZpwSpKmA).mp4 | audio: Audio Title - Audio Author (soundcloud, 1242868615).mp3.\n")
	audioOnly       = flag.Bool("audio", false, "Downloads only the audio, and removes the video.\n")
	videoOnly       = flag.Bool("video", false, "Downlods only the video, and removes the audio.\n")
	tiktokH265      = flag.Bool("tiktok-h265", false, "Downloads TikTok videos using h265 codec.\n")
	tiktokFullAudio = flag.Bool("tiktok-full", false, "Download the original sound used in a tiktok video.\n")
	dubbedAudio     = flag.Bool("dubbed-audio", false, "Downloads youtube audio dubbed, if present. Change the language using -language <ISO 639-1 format>.\n")
	metadataEmbed   = flag.Bool("metadata", false, "Don't embeds file metadata, if possible, to the download.\n")
	twitterGif      = flag.Bool("gif", true, "Convert twitter gifs to .gif.\n")
	apiOverride     = flag.String("api", "https://api.cobalt.tools", "Change the cobalt api url used. See others instances in https://instances.hyper.lol.\n")
	dubLang         = flag.String("language", "en", "Downloads dubbed youtube audio according to the language set following the ISO 639-1 format. Only takes effect if -dubbed-audio was passed as an argument.\n")
)

type Settings struct {
	Url                  string `json:"url"`             //Any URL from bilibili.com, instagram, pinterest, reddit, rutube, soundcloud, streamable, tiktok, tumblr, twitch clips, twitter/x, vimeo, vine archive, vk or youtube. Will be url encoded later.
	VideoCodec           string `json:"vCodec"`          //H264, AV1 or VP9, defaults to H264.
	VideoQuality         int    `json:"vQuality,string"` //144p to 2160p (4K), if not specified will default to 1080p.
	AudioCodec           string `json:"aFormat"`         //MP3, Opus, Ogg or Wav. If not specified will default to best.
	FilenamePattern      string `json:"filenamePattern"` //Classic, Basic, Pretty or Nerdy. Defaults to Pretty
	AudioOnly            bool   `json:"isAudioOnly"`     //Removes the video, downloads audio only. Default: false
	TikTokH265           bool   `json:"tiktokH265"`      //Changes whether 1080p h265 [tiktok] videos are preferred or not. Default: false
	FullTikTokAudio      bool   `json:"isTTFullAudio"`   //Enables download of original sound used in a tiktok video. Default: false
	VideoOnly            bool   `json:"isAudioMuted"`    //Downloads only the video, audio is muted/removed. Default: false
	DubbedYoutubeAudio   bool   `json:"dubLang"`         //Pass the User-Language HTTP header to use the dubbed audio of the respective language, must change according to user's preference, default is English (US). Uses ISO 639-1 standard.
	DisableVideoMetadata bool   `json:"disableMetadata"` //Removes file metadata. Default: false
	ConvertTwitterGifs   bool   `json:"twitterGif"`      //Changes whether twitter gifs are converted to .gif (Twitter gifs are usually stored in .mp4 format). Default: true
}

func main() {
	flag.Parse()
	validUrl, _ := regexp.MatchString(`[-a-zA-Z0-9@:%_+.~#?&/=]{2,256}\.[a-z]{2,4}\b(/[-a-zA-Z0-9@:%_+.~#?&/=]*)?`, *url)
	newDownload := Settings{}

	//Error checking
	if *url == "" || !validUrl {
		bs := path.Base(os.Args[0])
		fmt.Printf("invalid usage of retrobalt, use \"%v -help\" for usage", bs)
		return
	}
	newDownload.Url = *url

	if *videoQuality < 144 || *videoQuality > 2160 {
		fmt.Println("invalid video quality provided.")
		return
	}
	newDownload.VideoQuality = *videoQuality

	switch *videoCodec {
	case "vp9", "av1", "h264":
		newDownload.VideoCodec = *videoCodec
	default:
		fmt.Println("invalid video codec provided.")
		return
	}

	switch *audioCodec {
	case "best", "ogg", "wav", "opus", "mp3":
		newDownload.AudioCodec = *audioCodec
	default:
		fmt.Println("invalid audio codec provided.")
		return
	}

	switch *filenamePattern {
	case "classic", "basic", "nerdy", "pretty":
		newDownload.FilenamePattern = *filenamePattern
	default:
		fmt.Println("invalid file name pattern provided.")
		return
	}

	if *apiOverride == "" {
		fmt.Println("no api url was provided.")
		return
	}
	newDownload.AudioOnly = *audioOnly
	newDownload.VideoOnly = *videoOnly
	newDownload.TikTokH265 = *tiktokH265
	newDownload.FullTikTokAudio = *tiktokFullAudio
	newDownload.DubbedYoutubeAudio = *dubbedAudio
	newDownload.DisableVideoMetadata = *metadataEmbed
	newDownload.ConvertTwitterGifs = *twitterGif

	log.Println("[info] Everything sounds good, sending request to cobalt... ")

	//Requets to cobalt
	resp, err := doCobaltRequest(newDownload, *apiOverride)
	if err != nil {
		fmt.Println("[error] unable to get the final url:", err)
		return
	}
	
	info, err := parseMedia(resp.URL)
	if err != nil {
		fmt.Println("[error] unable to parse file:", err)
		return
	}
	a, err := downloadFile(*info)
	if err != nil {
		fmt.Println("[error] unable to download file:", err)
		return
	}
	log.Println(a)
}
