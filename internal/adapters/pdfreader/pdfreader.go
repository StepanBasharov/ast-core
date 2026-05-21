package pdfreader

import (
	"strings"

	"backend/pkg/log"

	"github.com/ledongthuc/pdf"
)

type PDFReader struct {
	log log.Logger
}

func NewPDFReader(log log.Logger) *PDFReader {
	return &PDFReader{
		log: log,
	}
}

func (p *PDFReader) Read(path string) (string, error) {
	file, reader, err := pdf.Open(path)
	if err != nil {
		return "", err
	}

	defer func() {
		if errClose := file.Close(); errClose != nil {
			p.log.Error(
				"error open file",
				log.FieldLogger{Key: "file", Value: path},
				log.FieldLogger{Key: "error", Value: errClose},
			)
		}
	}()

	var sb strings.Builder

	for i := 1; i <= reader.NumPage(); i++ {
		page := reader.Page(i)
		if page.V.IsNull() {
			continue
		}
		rows, err := page.GetTextByRow()
		if err != nil {
			continue
		}
		for _, row := range rows {
			for _, word := range row.Content {
				sb.WriteString(word.S)
				sb.WriteString(" ")
			}
			sb.WriteString("\n")
		}
	}

	return sb.String(), nil
}
