package geo

import (
	"qrcodegen/config"
	"qrcodegen/internal/usecase"

	"github.com/rs/zerolog/log"
)

func NewGeoResolver(cfg *config.Config) (usecase.GeoResolver, error) {
	if cfg.IPInfoToken != "" {
		log.Info().Msg("Using IPinfo Lite for Geo resolving")
		return newIPInfoLite(cfg.IPInfoToken, cfg.IPInfoHTTPTimeout), nil
	}
	log.Warn().Msg("No IPINFO_TOKEN; geo resolving disabled.")
	return &noOpResolver{}, nil
}

type noOpResolver struct{}

func (r *noOpResolver) Resolve(ip string) (string, string, bool) { return "", "", false }
