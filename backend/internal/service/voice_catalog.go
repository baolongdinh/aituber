package service

import (
	"fmt"
	"strings"
)

// VoiceMetadata defines voice characteristics and provider support
type VoiceMetadata struct {
	Key             string   `json:"key"`             // banmai, minhquang
	Label           string   `json:"label"`           // Ban Mai, Minh Quang
	Gender          string   `json:"gender"`          // male, female
	ProviderSupport []string `json:"providerSupport"` // ["fpt"], ["hub"], ["fpt","hub"]
	RefAudioFile    string   `json:"refAudioFile"`    // man_north_sound.mp3, women_north_sound.mp3
}

// VoiceCatalog provides voice metadata and resolution
type VoiceCatalog struct {
	voices map[string]VoiceMetadata
}

// NewVoiceCatalog creates and initializes the voice catalog
func NewVoiceCatalog() *VoiceCatalog {
	catalog := &VoiceCatalog{
		voices: make(map[string]VoiceMetadata),
	}

	// Initialize voice catalog with current and future voices
	voices := []VoiceMetadata{
		// Female voices - FPT only (2 voices)
		{
			Key:             "banmai",
			Label:           "Ban Mai",
			Gender:          "female",
			ProviderSupport: []string{"fpt", "hub"}, // Hub supports this voice
			RefAudioFile:    "women_north_sound.mp3",
		},
		{
			Key:             "leminh",
			Label:           "Lê Minh",
			Gender:          "female",
			ProviderSupport: []string{"fpt"}, // FPT only
			RefAudioFile:    "women_north_sound.mp3",
		},

		// Male voices - FPT only (2 voices)
		{
			Key:             "minhquang",
			Label:           "Minh Quang",
			Gender:          "male",
			ProviderSupport: []string{"fpt", "hub"}, // Hub supports this voice
			RefAudioFile:    "man_north_sound.mp3",
		},
		{
			Key:             "giahuy",
			Label:           "Gia Huy",
			Gender:          "male",
			ProviderSupport: []string{"fpt"}, // FPT only
			RefAudioFile:    "man_north_sound.mp3",
		},
	}

	for _, voice := range voices {
		catalog.voices[voice.Key] = voice
	}

	return catalog
}

// GetVoiceMetadata returns metadata for a given voice key
func (vc *VoiceCatalog) GetVoiceMetadata(voiceKey string) (VoiceMetadata, error) {
	voice, exists := vc.voices[voiceKey]
	if !exists {
		return VoiceMetadata{}, fmt.Errorf("voice '%s' not found in catalog", voiceKey)
	}
	return voice, nil
}

// GetVoicesByProvider returns all voices supported by a specific provider
func (vc *VoiceCatalog) GetVoicesByProvider(provider string) []VoiceMetadata {
	var voices []VoiceMetadata
	for _, voice := range vc.voices {
		for _, supportedProvider := range voice.ProviderSupport {
			if supportedProvider == provider {
				voices = append(voices, voice)
				break
			}
		}
	}
	return voices
}

// GetVoicesByGender returns all voices of a specific gender for a provider
func (vc *VoiceCatalog) GetVoicesByGender(provider, gender string) []VoiceMetadata {
	var voices []VoiceMetadata
	providerVoices := vc.GetVoicesByProvider(provider)
	for _, voice := range providerVoices {
		if voice.Gender == gender {
			voices = append(voices, voice)
		}
	}
	return voices
}

// GetRefAudioURL returns the public URL for a voice's reference audio file
func (vc *VoiceCatalog) GetRefAudioURL(baseURL, voiceKey string) (string, error) {
	voice, err := vc.GetVoiceMetadata(voiceKey)
	if err != nil {
		return "", err
	}

	// Ensure baseURL doesn't end with slash and voice file doesn't start with slash
	baseURL = strings.TrimSuffix(baseURL, "/")
	refAudioFile := strings.TrimPrefix(voice.RefAudioFile, "/")

	return fmt.Sprintf("%s/voice/%s", baseURL, refAudioFile), nil
}

// ValidateProvider checks if a voice supports the given provider
func (vc *VoiceCatalog) ValidateProvider(voiceKey, provider string) error {
	voice, err := vc.GetVoiceMetadata(voiceKey)
	if err != nil {
		return err
	}

	for _, supportedProvider := range voice.ProviderSupport {
		if supportedProvider == provider {
			return nil
		}
	}

	return fmt.Errorf("voice '%s' does not support provider '%s'. Supported: %v",
		voiceKey, provider, voice.ProviderSupport)
}

// GetAllVoices returns all voices in the catalog
func (vc *VoiceCatalog) GetAllVoices() []VoiceMetadata {
	var voices []VoiceMetadata
	for _, voice := range vc.voices {
		voices = append(voices, voice)
	}
	return voices
}

// IsSupportedProvider checks if a provider is valid
func (vc *VoiceCatalog) IsSupportedProvider(provider string) bool {
	supportedProviders := []string{"fpt", "hub"}
	for _, p := range supportedProviders {
		if p == provider {
			return true
		}
	}
	return false
}
