package qrcode

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image/png"
	"regexp"
	"strings"

	"github.com/jung-kurt/gofpdf"
	"github.com/quickqr/gqr"
	export "github.com/quickqr/gqr/export/image"
	"github.com/quickqr/gqr/export/image/shapes"
)

func normalizeHex(s string) (string, error) {
	s = strings.TrimSpace(strings.TrimPrefix(s, "#"))
	if ok, _ := regexp.MatchString("^[0-9a-fA-F]{6}$", s); !ok {
		return "", fmt.Errorf("invalid hex color: %q", s)
	}
	return strings.ToUpper(s), nil
}

func GeneratePNG(url, colorHex, bgHex string, smoothing float64) ([]byte, error) {
	fg, err := normalizeHex(colorHex)
	if err != nil {
		fg = "5EC8FF"
	}

	bg, err := normalizeHex(bgHex)
	if err != nil {
		bg = "FFFFFF"
	}

    radius := smoothing
    if radius < 0 {
        radius = 0
    }
    if radius > 0.5 {
        radius = 0.5
    }

    qr, err := gqr.NewWith(
		url,
		gqr.WithErrorCorrectionLevel(gqr.ErrorCorrectionHighest),
		gqr.WithEncodingMode(gqr.EncModeAuto),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create qr matrix: %w", err)
	}

    exp := export.NewExporter(
		export.WithImageSize(1024),
		export.WithQuietZone(48),
		export.WithModuleGap(0.14),
		export.WithBgColorHex("#"+bg),

        // apply smoothing to rounded corners
        export.WithModuleShape(shapes.RoundedModuleShape(radius, true)),
        export.WithFinderShape(shapes.RoundedFinderShape(radius)),

		export.WithGradient(
			export.GradientDirectionLTR,
			export.ParseFromHex("#"+fg),
			export.ParseFromHex("#"+fg),
		),
	)

	img := exp.Export(*qr)

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, fmt.Errorf("failed to encode png: %w", err)
	}
	return buf.Bytes(), nil
}

func GenerateSVG(url, colorHex, bgHex string, smoothing float64) ([]byte, error) {
	png, err := GeneratePNG(url, colorHex, bgHex, smoothing)
	if err != nil {
		return nil, err
	}
	b64 := base64.StdEncoding.EncodeToString(png)
	svg := fmt.Sprintf(
		`<svg xmlns="http://www.w3.org/2000/svg" width="1024" height="1024" viewBox="0 0 1024 1024"><image width="1024" height="1024" href="data:image/png;base64,%s"/></svg>`,
		b64,
	)
	return []byte(svg), nil
}

func GeneratePDF(url, colorHex, bgHex string, smoothing float64) ([]byte, error) {
	png, err := GeneratePNG(url, colorHex, bgHex, smoothing)
	if err != nil {
		return nil, err
	}

	pdf := gofpdf.NewCustom(&gofpdf.InitType{
		UnitStr: "mm",
		Size:    gofpdf.SizeType{Wd: 270, Ht: 270},
	})
	pdf.AddPage()

	opt := gofpdf.ImageOptions{ImageType: "PNG", ReadDpi: true}
	pdf.RegisterImageOptionsReader("qr", opt, bytes.NewReader(png))
	pdf.ImageOptions("qr", 0, 0, 270, 270, false, opt, 0, "")

	var out bytes.Buffer
	if err := pdf.Output(&out); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}
