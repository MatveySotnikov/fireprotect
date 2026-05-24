package internal

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"time"

	pb "github.com/MatveySotnikov/fireprotect/gen/pdfpb"
	"github.com/jung-kurt/gofpdf"
)

//go:embed TIMES.TTF
var timesFont []byte

type PdfServiceServer struct {
	pb.UnimplementedPdfServiceServer
}

func (s *PdfServiceServer) GenerateAct(ctx context.Context, req *pb.PdfRequest) (*pb.PdfResponse, error) {
	// Парсим дату
	calcDate, err := time.Parse(time.RFC3339, req.GetCalcDate())
	if err != nil {
		return nil, fmt.Errorf("invalid calc_date: %w", err)
	}

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Регистрируем шрифт из встроенных байт
	pdf.AddUTF8FontFromBytes("Times", "", timesFont)

	// Название
	pdf.SetFont("Times", "", 16)
	pdf.CellFormat(190, 10, "Акт расчёта огнезащитного состава", "", 1, "C", false, 0, "")
	pdf.Ln(4)

	// Данные пользователя
	pdf.SetFont("Times", "", 12)
	pdf.Cell(40, 8, "Заказчик:")
	pdf.Cell(0, 8, req.GetUserName())
	pdf.Ln(7)
	pdf.Cell(40, 8, "Email:")
	pdf.Cell(0, 8, req.GetUserEmail())
	pdf.Ln(10)

	// Исходные данные
	pdf.SetFont("Times", "", 12)
	pdf.Cell(0, 8, "Исходные данные")
	pdf.Ln(8)
	pdf.SetFont("Times", "", 12)
	addRow(pdf, "Площадь поверхности, м²", fmt.Sprintf("%.2f", req.GetArea()))
	addRow(pdf, "Нормативный расход на слой, кг/м²", fmt.Sprintf("%.2f", req.GetNormativeRate()))
	addRow(pdf, "Количество слоёв", fmt.Sprintf("%d", req.GetLayers()))
	addRow(pdf, "Угол уклона кровли, °", fmt.Sprintf("%.1f", req.GetSlopeAngle()))
	addRow(pdf, "Коэффициент потерь", fmt.Sprintf("%.2f", req.GetLossFactor()))
	addRow(pdf, "Плотность состава, кг/л", fmt.Sprintf("%.2f", req.GetDensity()))
	pdf.Ln(5)

	// Результаты
	pdf.SetFont("Times", "", 12)
	pdf.Cell(0, 8, "Результаты расчёта")
	pdf.Ln(8)
	pdf.SetFont("Times", "", 12)
	addRow(pdf, "Общая масса состава, кг", fmt.Sprintf("%.2f", req.GetTotalMass()))
	addRow(pdf, "Общий объём состава, л", fmt.Sprintf("%.2f", req.GetTotalVolume()))
	pdf.Ln(5)

	// Дата
	pdf.Cell(0, 8, fmt.Sprintf("Дата расчёта: %s", calcDate.Format("02.01.2006 15:04:05")))
	pdf.Ln(15)

	// Подпись
	pdf.Cell(0, 8, "Расчёт выполнен автоматически сервисом FireProtect")
	pdf.Ln(5)
	pdf.Cell(0, 8, "Подпись: ____________________")

	var buf bytes.Buffer
	err = pdf.Output(&buf)
	if err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}

	return &pb.PdfResponse{
		PdfData: buf.Bytes(),
	}, nil
}

func addRow(pdf *gofpdf.Fpdf, label, value string) {
	pdf.Cell(100, 7, label)
	pdf.Cell(0, 7, value)
	pdf.Ln(6)
}
