/*
 * @Author: kamalyes 501893067@qq.com
 * @Date: 2024-12-13 09:55:55
 * @LastEditors: kamalyes 501893067@qq.com
 * @LastEditTime: 2025-01-07 17:16:55
 * @FilePath: \go-toolbox\pkg\imgix\drawer.go
 * @Description:
 *
 * Copyright (c) 2024 by kamalyes, All Rights Reserved.
 */
package imgix

import (
	"bytes"
	"fmt"
	"image"
	"log"
	"math"
	"sync"

	"github.com/disintegration/imaging"
	"github.com/fogleman/gg"
	"github.com/HitGameAI/apex-toolbox/pkg/mathx"
	"github.com/HitGameAI/apex-toolbox/pkg/syncx"
)

// GraphicsRenderer 结构体用于绘制面部特征
type GraphicsRenderer struct {
	GgCtx       *gg.Context  // Gg 上下文
	DashOptions DashOptions  // 虚实线扩展
	bufferPool  *sync.Pool   // bufferPool 指针
	mu          sync.RWMutex // 读写锁
}

type DashStyle float64

type DashOptions struct {
	dashLength DashStyle // 虚线段长度
	gapLength  DashStyle // 虚线间隔长度
	lineWidth  DashStyle // 线条宽度
}

const (
	defaultDashLength = 3
	defaultGapLength  = 6
	defaultLineWidth  = 2
)

// NewDashOptions 创建一个新的 DashOptions 实例
// 如果提供了自定义的段长度和间隔长度，则使用它们；否则使用默认值
func NewDashOptions(dashLength, gapLength, lineWidth DashStyle) DashOptions {
	// 默认值
	dashLength = mathx.IfLeZero(dashLength, defaultDashLength)
	gapLength = mathx.IfLeZero(gapLength, defaultGapLength)
	lineWidth = mathx.IfLeZero(lineWidth, defaultLineWidth)

	return DashOptions{
		dashLength: dashLength,
		gapLength:  gapLength,
		lineWidth:  lineWidth,
	}
}

// SetDashed 设置线条
// @param dashes
func (g *GraphicsRenderer) SetDashed(dashes ...float64) {
	g.DrawWithStroke(func() {
		g.GgCtx.SetDash(dashes...)
	}, false)
}

// DashLength 返回虚线段长度的 float64 表示
func (d *DashOptions) DashLength() DashStyle {
	return d.dashLength
}

// GapLength 返回虚线间隔长度的 float64 表示
func (d *DashOptions) GapLength() DashStyle {
	return d.gapLength
}

// 获取 LineWidth 的方法
func (g *DashOptions) LineWidth() DashStyle {
	return g.lineWidth
}

// 获取 DashOptions 的方法
func (g *GraphicsRenderer) GetDashOptions() DashOptions {
	return g.DashOptions
}

// ImageLTRB 表示图像的边界框
type ImageLTRB struct {
	Left   float64 // 左边界
	Top    float64 // 上边界
	Right  float64 // 右边界
	Bottom float64 // 下边界
}

// Width 计算边界框的宽度
func (ltrb ImageLTRB) Width() float64 {
	return ltrb.Right - ltrb.Left
}

// Height 计算边界框的高度
func (ltrb ImageLTRB) Height() float64 {
	return ltrb.Bottom - ltrb.Top
}

// Center 返回边界框的中心点
func (ltrb ImageLTRB) Center() (float64, float64) {
	centerX := ltrb.Left + ltrb.Width()/2
	centerY := ltrb.Top + ltrb.Height()/2
	return centerX, centerY
}

// Contains 检查一个点是否在边界框内
func (ltrb ImageLTRB) Contains(x, y float64) bool {
	return x >= ltrb.Left && x <= ltrb.Right && y >= ltrb.Top && y <= ltrb.Bottom
}

// String 返回边界框的字符串表示
func (ltrb ImageLTRB) String() string {
	return fmt.Sprintf("ImageLTRB(Left: %f, Top: %f, Right: %f, Bottom: %f)", ltrb.Left, ltrb.Top, ltrb.Right, ltrb.Bottom)
}

// NewGraphicsRenderer 创建一个新的 GraphicsRenderer 实例
// @param ctx gg.Context 的指针，用于绘制操作
// @param dashOptions
// @return *GraphicsRenderer 返回一个新的 GraphicsRenderer 实例
func NewGraphicsRenderer(ctx *gg.Context, dashOptions ...DashOptions) *GraphicsRenderer {
	defaultDashOptions := NewDashOptions(defaultDashLength, defaultGapLength, defaultLineWidth)
	if len(dashOptions) > 0 {
		defaultDashOptions = dashOptions[0]
	}
	// 初始化 bufferPool
	bufferPool := &sync.Pool{
		New: func() interface{} {
			return new(bytes.Buffer)
		},
	}
	renderer := GraphicsRenderer{
		GgCtx:       ctx,
		DashOptions: defaultDashOptions,
		bufferPool:  bufferPool,
	}
	return &renderer
}

// DrawWithStroke 是一个通用的绘图函数，接受一个绘图操作的函数作为参数
func (g *GraphicsRenderer) DrawWithStroke(drawFunc func(), isStroke bool) {
	syncx.WithLock(&g.mu, func() {
		drawFunc() // 执行绘图操作
		if isStroke {
			g.GgCtx.Stroke() // 在绘图操作完成后统一调用 Stroke
		}
	})
}

// UseDefaultDashed 使用默认虚线
func (g *GraphicsRenderer) UseDefaultDashed() {
	g.DrawWithStroke(func() {
		g.GgCtx.SetDash(float64(g.DashOptions.DashLength()), float64(g.DashOptions.GapLength()), 0.0)
	}, false)
}

// UseSolidLine 使用实线
func (g *GraphicsRenderer) UseSolidLine() {
	g.DrawWithStroke(func() {
		g.GgCtx.SetDash() // 恢复为实线样式
	}, false)
}

// 缩放图扩展
type ResizeImgOptions struct {
	Width  int
	Height int
	Filter imaging.ResampleFilter
}

// ResizeImage 缩放图片
// @param img: 要缩放的原始图像（image.Image类型）
// @param resizeImgOptions: 缩放选项（*CropImgOptions类型），包含缩放的宽高
// return
//
//	返回缩放后的图像（image.Image类型）
func ResizeImage(img image.Image, resizeImgOptions *ResizeImgOptions) image.Image {
	// 检查是否需要缩放，并确保 Width 和 Height 大于 0
	if resizeImgOptions != nil && resizeImgOptions.Width > 0 && resizeImgOptions.Height > 0 {
		img = imaging.Resize(img, resizeImgOptions.Width, resizeImgOptions.Height, resizeImgOptions.Filter)
	}
	return img
}

// CropImgOptions 定义裁剪图像的选项
type CropImgOptions struct {
	MinWidth  int // 最小裁剪宽度
	MinHeight int // 最小裁剪高度
	MaxWidth  int // 最大裁剪宽度
	MaxHeight int // 最大裁剪高度
}

// CropImage 裁剪图片
// @param img: 要裁剪的原始图像（image.Image类型）
// @param cropImgOptions: 裁剪选项（*CropImgOptions类型），包含裁剪的最小和最大宽高
// return
//
//	返回裁剪后的图像（image.Image类型）
func CropImage(img image.Image, cropImgOptions *CropImgOptions) image.Image {
	// 检查裁剪选项是否不为 nil 并确保 MaxWidth 和 MaxHeight 大于 0
	if cropImgOptions != nil && cropImgOptions.MaxWidth > 0 && cropImgOptions.MaxHeight > 0 {
		// 创建裁剪区域的矩形
		cropRect := image.Rect(cropImgOptions.MinWidth, cropImgOptions.MinHeight, cropImgOptions.MaxWidth, cropImgOptions.MaxHeight)
		// 使用 imaging 库裁剪图像
		img = imaging.Crop(img, cropRect)
	}
	return img // 返回裁剪后的图像
}

// SaveImage 存储图片
// @param img 图像
// @param name 文件名名称
// @param quality 图片质量
// @param imageFormat 图片类型 (目前仅支持PNG/JPG/JPEG)
func (g *GraphicsRenderer) SaveImage(img image.Image, name string, quality int, imageFormat ImageFormat) error {
	// 从池中获取一个 bytes.Buffer，并在完成后放回
	imgBuffer := g.bufferPool.Get().(*bytes.Buffer)
	defer g.bufferPool.Put(imgBuffer)

	// 清空缓冲区并获取图像
	imgBuffer.Reset()
	// 写入图像
	if err := WriterImage(img, quality, imageFormat, imgBuffer); err != nil {
		return err
	}

	// 保存生成的图像
	imageFileName := fmt.Sprintf("%s.%s", name, imageFormat.String())
	return SaveBufToImageFile(imgBuffer, imageFileName, imageFormat)
}

// DrawLineXYLineWidth 绘制线条
// @param startX 起始X坐标
// @param startY 起始Y坐标
// @param endX 结束X坐标
// @param endY 结束Y坐标
// @param lineWidth 线条宽度
func (g *GraphicsRenderer) DrawLineXYLineWidth(startX, startY, endX, endY float64) {
	g.DrawWithStroke(func() {
		g.GgCtx.DrawLine(startX, startY, endX, endY)
	}, true)

}

// DrawCurvedLine 绘制带有弯曲度的线条
// @param start 起始点
// @param end 结束点
// @param control 控制点
func (g *GraphicsRenderer) DrawCurvedLine(start, end, control *gg.Point) {
	g.DrawWithStroke(func() {
		g.GgCtx.MoveTo(start.X, start.Y)
		g.GgCtx.QuadraticTo(control.X, control.Y, end.X, end.Y)
	}, true)
}

// DelayLine 表示每条延长线的绘制状态和是否使用虚线
type DelayLine struct {
	Enabled      bool    // 是否绘制该延长线
	Dash         bool    // 是否使用虚线
	extendLength float64 // 延长长度
}

// RectangleExtensionFlags 用于控制矩形每个角的延长线绘制状态
type RectangleExtensions struct {
	RightTop    DelayLine // 右上角的上方延长线
	RightBottom DelayLine // 右下角的下方延长线
	LeftTop     DelayLine // 左上角的上方延长线
	LeftBottom  DelayLine // 左下角的下方延长线
	TopLeft     DelayLine // 上方的左侧延长线
	TopRight    DelayLine // 上方的右侧延长线
	BottomLeft  DelayLine // 下方的左侧延长线
	BottomRight DelayLine // 下方的右侧延长线
}

// NewRectangleExtensions 创建一个新的 RectangleExtensions 实例，所有字段默认为不绘制且不使用虚线
func NewRectangleExtensions() *RectangleExtensions {
	return &RectangleExtensions{
		RightTop:    DelayLine{},
		RightBottom: DelayLine{},
		LeftTop:     DelayLine{},
		LeftBottom:  DelayLine{},
		TopLeft:     DelayLine{},
		TopRight:    DelayLine{},
		BottomLeft:  DelayLine{},
		BottomRight: DelayLine{},
	}
}

// NewRectangle 创建一个新的矩形，默认延长线不绘制
func NewRectangle(topLeft, bottomRight *gg.Point, rectangleExtensions ...*RectangleExtensions) Rectangle {
	rectExtensions := NewRectangleExtensions()
	if len(rectangleExtensions) > 0 {
		rectExtensions = rectangleExtensions[0]
	}
	return Rectangle{
		TopLeft:     topLeft,
		BottomRight: bottomRight,
		Extensions:  rectExtensions, // 默认不绘制延长线
	}
}

// Rectangle 表示一个矩形，包含左上角和右下角的坐标，以及控制延长线绘制的选项
type Rectangle struct {
	TopLeft     *gg.Point            // 矩形的左上角坐标
	BottomRight *gg.Point            // 矩形的右下角坐标
	Extensions  *RectangleExtensions // 控制延长线的绘制状态，默认为不绘制
}

// DrawRectangle 绘制矩形框，并可选地绘制延长线
// @param rect 矩形的左上角和右下角
// @param extendLength 延长线的长度，单位与坐标相同
func (g *GraphicsRenderer) DrawRectangle(rect Rectangle, extendLength float64) {
	rectExtensions := NewRectangleExtensions()
	if rect.Extensions == nil {
		rect.Extensions = rectExtensions
	}
	g.DrawWithStroke(func() {
		width := rect.BottomRight.X - rect.TopLeft.X
		height := rect.BottomRight.Y - rect.TopLeft.Y
		g.GgCtx.DrawRectangle(rect.TopLeft.X, rect.TopLeft.Y, width, height)
	}, true)
	// 绘制延长线
	g.drawExtensionLines(rect, extendLength)
}

// setRectangleLineStyle 根据给定的 Extension 设置线条样式
func (g *GraphicsRenderer) setRectangleLineStyle(extension DelayLine) {
	if extension.Dash {
		g.UseDefaultDashed() // 设置虚线样式
	} else {
		g.UseSolidLine() // 设置实线样式
	}
}

// getRectangleExtendLength 获取延长线长度
func (g *GraphicsRenderer) getRectangleExtendLength(parentExtendLength, sonExtendLength float64) float64 {
	if sonExtendLength > 0 {
		return sonExtendLength
	}
	return parentExtendLength
}

// drawExtensionLines 绘制延长线
func (g *GraphicsRenderer) drawExtensionLines(rect Rectangle, parentLength float64) {
	// 根据 ExtensionFlags 决定是否绘制延长线

	// 左上角的上方延长线
	if rect.Extensions.TopLeft.Enabled {
		g.setRectangleLineStyle(rect.Extensions.TopLeft)
		extendLength := g.getRectangleExtendLength(parentLength, rect.Extensions.TopLeft.extendLength)
		g.DrawLine(rect.TopLeft, OffsetPointY(rect.TopLeft, extendLength, Subtract))
	}

	// 右上角的上方延长线
	if rect.Extensions.TopRight.Enabled {
		g.setRectangleLineStyle(rect.Extensions.TopRight)
		extendLength := g.getRectangleExtendLength(parentLength, rect.Extensions.TopRight.extendLength)
		g.DrawLine(&gg.Point{X: rect.BottomRight.X, Y: rect.TopLeft.Y}, &gg.Point{X: rect.BottomRight.X, Y: rect.TopLeft.Y - extendLength})
	}

	// 左下角的下方延长线
	if rect.Extensions.BottomLeft.Enabled {
		g.setRectangleLineStyle(rect.Extensions.BottomLeft)
		extendLength := g.getRectangleExtendLength(parentLength, rect.Extensions.BottomLeft.extendLength)
		g.DrawLine(&gg.Point{X: rect.TopLeft.X, Y: rect.BottomRight.Y}, &gg.Point{X: rect.TopLeft.X, Y: rect.BottomRight.Y + extendLength})
	}

	// 右下角的下方延长线
	if rect.Extensions.BottomRight.Enabled {
		g.setRectangleLineStyle(rect.Extensions.BottomRight)
		extendLength := g.getRectangleExtendLength(parentLength, rect.Extensions.BottomRight.extendLength)
		g.DrawLine(rect.BottomRight, OffsetPointY(rect.BottomRight, extendLength, Add))
	}

	// 左上角的左侧延长线
	if rect.Extensions.LeftTop.Enabled {
		g.setRectangleLineStyle(rect.Extensions.LeftTop)
		extendLength := g.getRectangleExtendLength(parentLength, rect.Extensions.LeftTop.extendLength)
		g.DrawLine(rect.TopLeft, OffsetPointX(rect.TopLeft, extendLength, Subtract))
	}

	// 左下角的左侧延长线
	if rect.Extensions.LeftBottom.Enabled {
		g.setRectangleLineStyle(rect.Extensions.LeftBottom)
		extendLength := g.getRectangleExtendLength(parentLength, rect.Extensions.LeftBottom.extendLength)
		g.DrawLine(&gg.Point{X: rect.TopLeft.X, Y: rect.BottomRight.Y}, &gg.Point{X: rect.TopLeft.X - extendLength, Y: rect.BottomRight.Y})
	}

	// 右上角的右侧延长线
	if rect.Extensions.RightTop.Enabled {
		g.setRectangleLineStyle(rect.Extensions.RightTop)
		extendLength := g.getRectangleExtendLength(parentLength, rect.Extensions.RightTop.extendLength)
		g.DrawLine(&gg.Point{X: rect.BottomRight.X, Y: rect.TopLeft.Y}, &gg.Point{X: rect.BottomRight.X + extendLength, Y: rect.TopLeft.Y})
	}

	// 右下角的右侧延长线
	if rect.Extensions.RightBottom.Enabled {
		g.setRectangleLineStyle(rect.Extensions.RightBottom)
		extendLength := g.getRectangleExtendLength(parentLength, rect.Extensions.RightBottom.extendLength)
		g.DrawLine(rect.BottomRight, &gg.Point{X: rect.BottomRight.X + extendLength, Y: rect.BottomRight.Y})
	}
}

// DrawPolygon 绘制一个多边形
// @param points 多边形顶点的切片
func (g *GraphicsRenderer) DrawPolygon(points []gg.Point) {
	g.DrawWithStroke(func() {
		if len(points) > 0 {
			g.GgCtx.MoveTo(points[0].X, points[0].Y)
			for _, point := range points[1:] {
				g.GgCtx.LineTo(point.X, point.Y)
			}
			g.GgCtx.ClosePath()
		}
	}, true)
}

// VerticalLine 表示竖线的参数
type VerticalLine struct {
	X       float64 // 竖线的X坐标
	TopY    float64 // 竖线的顶部Y坐标
	BottomY float64 // 竖线的底部Y坐标
}

// HorizontalLine 表示横线的参数
type HorizontalLine struct {
	Y      float64 // 横线的Y坐标
	LeftX  float64 // 横线的左侧X坐标
	RightX float64 // 横线的右侧X坐标
}

// DrawVerticalLine 从面部顶部到底部绘制竖线
// @param line 竖线的参数
func (g *GraphicsRenderer) DrawVerticalLine(line VerticalLine) {
	g.DrawWithStroke(func() {
		g.GgCtx.DrawLine(line.X, line.TopY, line.X, line.BottomY)
	}, true)
}

// DrawHorizontalLine 从面部左侧到右侧绘制横线
// @param line 横线的参数
func (g *GraphicsRenderer) DrawHorizontalLine(line HorizontalLine) {
	g.DrawWithStroke(func() {
		g.GgCtx.DrawLine(line.LeftX, line.Y, line.RightX, line.Y)
	}, true)
}

// DrawLine 绘制从 startPoint 到 endPoint 的线条。
// @param startPoint 线条的起始
// @param endPoint: 线条的结束点
func (g *GraphicsRenderer) DrawLine(startPoint, endPoint *gg.Point) {
	g.DrawWithStroke(func() {
		g.GgCtx.DrawLine(startPoint.X, startPoint.Y, endPoint.X, endPoint.Y)
	}, true)
}

// DrawCircle 绘制一个圆形
// @param point 圆心的X、Y坐标
// @param radius 圆的半径
func (g *GraphicsRenderer) DrawCircle(point *gg.Point, radius float64) {
	g.DrawWithStroke(func() {
		g.GgCtx.DrawCircle(point.X, point.Y, radius)
	}, true)
}

// DrawPoint 绘制一个圆点
// @param point 圆心的X、Y坐标
// @param radius 圆的半径
func (g *GraphicsRenderer) DrawPoint(point *gg.Point, radius float64) {
	g.DrawWithStroke(func() {
		g.GgCtx.DrawPoint(point.X, point.Y, radius)
	}, true)
}

// DrawArc 绘制从 startPoint 到 endPoint 的弧形。
// 控制点自动计算为起始点和结束点的中点向上偏移一定的高度。
// @param startPoint 线条的起始
// @param endPoint: 线条的结束点
// @param offset 控制点的偏移量
// @param clockwise 为 true 表示顺时针，false 表示逆时针。
func (g *GraphicsRenderer) DrawArc(startPoint, endPoint *gg.Point, offset float64, clockwise bool) {
	// 计算中点
	midPoint := &gg.Point{
		X: (startPoint.X + endPoint.X) / 2,
		Y: (startPoint.Y + endPoint.Y) / 2,
	}

	// 计算控制点，向上偏移 offset
	controlPoint := &gg.Point{
		X: midPoint.X - offset,
		Y: midPoint.Y - offset,
	}

	// 计算弧的起始角度和结束角度
	startAngle := math.Atan2(float64(controlPoint.Y-startPoint.Y), float64(controlPoint.X-startPoint.X))
	endAngle := math.Atan2(float64(controlPoint.Y-endPoint.Y), float64(controlPoint.X-endPoint.X))

	// 计算半径
	radius := math.Hypot(float64(endPoint.X-startPoint.X)/2, float64(endPoint.Y-startPoint.Y)/2)

	// 如果是顺时针绘制，确保角度顺序正确
	if clockwise {
		if startAngle < endAngle {
			endAngle -= 2 * math.Pi
		}
	} else {
		if startAngle > endAngle {
			endAngle += 2 * math.Pi
		}
	}

	g.DrawWithStroke(func() {
		// 使用 DrawArc 方法绘制弧形
		g.GgCtx.DrawArc(midPoint.X, midPoint.Y, radius, startAngle, endAngle)
	}, true)
}

// DrawEllipse 绘制一个椭圆
// @param point 椭圆中心的X、Y坐标
// @param width 椭圆的宽度
// @param height 椭圆的高度
func (g *GraphicsRenderer) DrawEllipse(point *gg.Point, width, height float64) {
	g.DrawWithStroke(func() {
		g.GgCtx.DrawEllipse(point.X, point.Y, width/2, height/2)
	}, true)
}

// DrawTriangle 绘制一个三角形
// @param x1 第一个顶点的X坐标
// @param y1 第一个顶点的Y坐标
// @param x2 第二个顶点的X坐标
// @param y2 第二个顶点的Y坐标
// @param x3 第三个顶点的X坐标
// @param y3 第三个顶点的Y坐标
func (g *GraphicsRenderer) DrawTriangle(x1, y1, x2, y2, x3, y3 float64) {
	g.DrawWithStroke(func() {
		g.GgCtx.MoveTo(x1, y1) // 移动到第一个顶点
		g.GgCtx.LineTo(x2, y2) // 连接到第二个顶点
		g.GgCtx.LineTo(x3, y3) // 连接到第三个顶点
		g.GgCtx.ClosePath()    // 关闭路径形成三角形
	}, true) // 默认调用 Stroke
}

// LineSegment 表示线段的参数
type LineSegment struct {
	StartCoord    *gg.Point // 起始坐标
	EndCoord      *gg.Point // 结束坐标
	TextOffsetX   float64   // 文本X偏移量
	TextOffsetY   float64   // 文本Y偏移量
	Texts         []string  // 文本内容
	DrawLine      bool      // 是否绘制线条
	SkipDrawTexts bool      // 跳过绘制文字
}

// DrawCenteredMultiLine 在多个指定的线段内绘制多组文本
// @param lines 线段的参数数组
func (g *GraphicsRenderer) DrawCenteredMultiLine(lines []LineSegment) {
	log.Println("Starting DrawCenteredMultiLine")

	for i, line := range lines {
		log.Printf("Drawing line group %d", i)
		startX, endX := line.StartCoord.X, line.EndCoord.X
		startY, endY := line.StartCoord.Y, line.EndCoord.Y

		// 计算线段的长度和方向
		deltaX := endX - startX
		deltaY := endY - startY
		length := math.Sqrt(deltaX*deltaX + deltaY*deltaY) // 线段的长度

		// 若DrawLine为True绘制线条
		if line.DrawLine {
			g.DrawWithStroke(func() {
				g.GgCtx.DrawLine(startX, startY, endX, endY)
			}, true)
		}

		// 若SkipDrawTexts为True绘制线条即跳过
		if line.SkipDrawTexts {
			continue
		}

		// 计算文本的起始位置
		posX := startX + deltaX/2 + line.TextOffsetX
		posY := startY + deltaY/2 + line.TextOffsetY

		// 计算文本的高度和间隔
		var textHeight float64
		for j, text := range line.Texts {
			// 测量文本高度
			_, height := g.GgCtx.MeasureString(text)
			textHeight = height * 1.6 // 设定文本之间的间隔

			// 根据线段方向调整文本的 Y 偏移
			var textOffsetX, textOffsetY float64
			if deltaX == 0 { // 竖线
				textOffsetX = posX // X 坐标固定
				// 计算 Y 偏移，确保文本垂直对齐
				textOffsetY = posY - (textHeight * float64(len(line.Texts)-1) / 2) + textHeight*float64(j)
			} else if deltaY == 0 { // 横线
				textOffsetX = posX
				textOffsetY = posY + textHeight/2 + textHeight*float64(j)
			} else { // 斜线
				// 计算斜线的偏移量
				offsetY := (textHeight * deltaY / length) * float64(len(line.Texts)) / 2
				textOffsetX = posX
				textOffsetY = posY - offsetY + textHeight*float64(j) - textHeight/2
			}

			// 如果是第一条文本，添加额外的偏移
			if j == 0 {
				textOffsetY += 5
			}

			// 绘制文本
			g.GgCtx.DrawStringAnchored(text, textOffsetX, textOffsetY, 0.5, 0.5)
		}
	}

	log.Println("Finished DrawCenteredMultiLine")
}

// VerticalEdge 结构体表示一个垂直边，包括顶点和底点
type VerticalEdge struct {
	Nadir  *gg.Point // 底点，表示边的下端点
	Vertex *gg.Point // 顶点，表示边的上端点
}

// HorizontalEdge 结构体表示一个水平边，包括最左边和最右边的点
type HorizontalEdge struct {
	LeftMost  *gg.Point // 最左边的点，表示边的左端点
	RightMost *gg.Point // 最右边的点，表示边的右端点
}

// Coordinates 结构体表示边界框的坐标，包含四条边
type Coordinates struct {
	Top    HorizontalEdge // 上边，包含最左边和最右边的点
	Bottom HorizontalEdge // 下边，包含最左边和最右边的点
	Left   VerticalEdge   // 左边，包含顶点和底点
	Right  VerticalEdge   // 右边，包含顶点和底点
}

// CleanCoordinates 函数接受一个 map[string]Point，返回清洗后的边界框坐标
func CleanCoordinates(pointsMap map[string]gg.Point) Coordinates {
	if len(pointsMap) == 0 {
		return Coordinates{}
	}

	// 初始化最小值和最大值
	minX := math.Inf(1)
	maxX := math.Inf(-1)
	minY := math.Inf(1)
	maxY := math.Inf(-1)

	// 遍历所有点，找到最小和最大坐标
	for _, p := range pointsMap {
		if p.X < minX {
			minX = p.X
		}
		if p.X > maxX {
			maxX = p.X
		}
		if p.Y < minY {
			minY = p.Y
		}
		if p.Y > maxY {
			maxY = p.Y
		}
	}

	// 创建边界框的四条边
	top := HorizontalEdge{
		LeftMost:  &gg.Point{X: minX, Y: maxY},
		RightMost: &gg.Point{X: maxX, Y: maxY},
	}
	bottom := HorizontalEdge{
		LeftMost:  &gg.Point{X: minX, Y: minY},
		RightMost: &gg.Point{X: maxX, Y: minY},
	}
	left := VerticalEdge{
		Nadir:  &gg.Point{X: minX, Y: minY},
		Vertex: &gg.Point{X: minX, Y: maxY},
	}
	right := VerticalEdge{
		Nadir:  &gg.Point{X: maxX, Y: minY},
		Vertex: &gg.Point{X: maxX, Y: maxY},
	}

	// 返回构建的 Coordinates
	return Coordinates{
		Top:    top,
		Bottom: bottom,
		Left:   left,
		Right:  right,
	}
}

// GetLTRB 获取关键点各部分的点横纵坐标最大小值，即：left，top，right，bottom
// @param features 面部特征点的映射
// @return ImageLTRB 返回包含边界信息的结构体
func GetLTRB(features map[string]gg.Point) ImageLTRB {
	defaultCoordinates := 0.0
	maxFloat := float64(^uint(0) >> 1) // 最大的 float64 值
	minFloat := -maxFloat              // 最小的 float64 值

	left, top, right, bottom := maxFloat, maxFloat, minFloat, minFloat

	for _, coord := range features {
		UpdateBounds(coord.X, coord.Y, &left, &top, &right, &bottom)
	}

	// 使用默认值替代未更新的边界值
	if left == maxFloat {
		left = defaultCoordinates
	}
	if top == maxFloat {
		top = defaultCoordinates
	}
	if right == minFloat {
		right = defaultCoordinates
	}
	if bottom == minFloat {
		bottom = defaultCoordinates
	}

	return ImageLTRB{Left: left, Top: top, Right: right, Bottom: bottom}
}

// UpdateBounds 更新边界值
// @param x 当前坐标的X值
// @param y 当前坐标的Y值
// @param left 左边界的指针
// @param top 上边界的指针
// @param right 右边界的指针
// @param bottom 下边界的指针
func UpdateBounds(x, y float64, left, top, right, bottom *float64) {
	if x < *left {
		*left = x
	}
	if y < *top {
		*top = y
	}
	if x > *right {
		*right = x
	}
	if y > *bottom {
		*bottom = y
	}
}

// CalculateControlPoint 计算控制点以实现弯曲效果
// @param startPoint 起始点
// @param endPoint 结束点
// @param offsetX 坐标X偏移量
// @param offsetY 坐标Y偏移量
// @return *gg.Point 返回计算得到的控制点
func CalculateControlPoint(startPoint, endPoint *gg.Point, offsetX, offsetY float64) *gg.Point {
	midX := (startPoint.X + endPoint.X) / 2
	midY := (startPoint.Y + endPoint.Y) / 2
	return &gg.Point{X: midX + offsetX, Y: midY + offsetY}
}

// CalculateFractionPointMode 类型定义，用于表示计算模式
type CalculateFractionPointMode int

const (
	Add CalculateFractionPointMode = iota
	Subtract
	Multiply
	Divide
)

// CalculateFractionPoint 计算任意分数的坐标
// @param startPoint 起始点
// @param endPoint 结束点
// @param fraction 计算的分数
// @param mode 计算模式（加、减、乘、除）
// @return *gg.Point 返回计算得到的坐标
func CalculateFractionPoint(startPoint, endPoint *gg.Point, fraction float64, mode CalculateFractionPointMode) *gg.Point {
	var x, y float64

	// 计算差值
	deltaX := endPoint.X - startPoint.X
	deltaY := endPoint.Y - startPoint.Y

	switch mode {
	case Add:
		x = deltaX + fraction
		y = deltaY + fraction
	case Subtract:
		x = deltaX - fraction
		y = deltaY - fraction
	case Multiply:
		x = deltaX * fraction
		y = deltaY * fraction
	case Divide:
		if fraction != 0 {
			x = deltaX / fraction
			y = deltaY / fraction
		} else {
			x = deltaX // 避免除以零
			y = deltaY
		}
	default:
		x = startPoint.X
		y = startPoint.Y
	}

	return &gg.Point{
		X: x,
		Y: y,
	}
}

// 定义轴的枚举类型
type AxisPointMode int

const (
	AxisX AxisPointMode = iota // X轴
	AxisY                      // Y轴
)

// 定义计算模式
type CalculateMode int

const (
	CalculateMax CalculateMode = iota // 最大值模式
	CalculateMin                      // 最小值模式
	CalculateAvg                      // 平均值模式
)

// CalculatePoint 返回两个点中在指定维度上较大的或较小的点
// @param a 第一个点
// @param b 第二个点
// @param calculateMode 指定是取最大点还是最小点（CalculateMode 枚举类型）
// @param axisMode 指定比较的轴（AxisPointMode 枚举类型）
// @return 返回在指定维度上较大的或较小的点
func CalculatePoint(a, b *gg.Point, calculateMode CalculateMode, axisMode AxisPointMode) *gg.Point {
	switch axisMode {
	case AxisY: // 如果比较Y轴
		if (calculateMode == CalculateMax && a.Y > b.Y) || (calculateMode == CalculateMin && a.Y < b.Y) {
			return a // 返回a点
		}
		return b // 返回b点
	case AxisX: // 默认比较X轴
		fallthrough // 允许使用fallthrough以简化代码
	default:
		if (calculateMode == CalculateMax && a.X > b.X) || (calculateMode == CalculateMin && a.X < b.X) {
			return a // 返回a点
		}
		return b // 返回b点
	}
}

// CalculateMultiplePoints 计算多个点中的最大或最小点
// @param points 要比较的点的切片
// @param calculateMode 指定是取最大点还是最小点（PointMaxMin 枚举类型）
// @param axisMode 指定比较的轴（AxisPointMode 枚举类型、默认X）
// @return 返回在指定维度上较大的或较小的点
func CalculateMultiplePoints(points []*gg.Point, calculateMode CalculateMode, axisMode ...AxisPointMode) *gg.Point {
	// 如果模式是平均值，直接计算平均点
	if calculateMode == CalculateAvg {
		sumX, sumY := 0.0, 0.0
		for _, point := range points {
			sumX += point.X
			sumY += point.Y
		}
		return &gg.Point{
			X: sumX / float64(len(points)),
			Y: sumY / float64(len(points)),
		}
	}

	// 如果是比大小就需要校验数组长度是否正确
	if len(points) == 0 {
		return nil // 如果没有点，返回nil
	}
	var axis = AxisX
	if len(axisMode) > 0 {
		axis = axisMode[0]
	}

	// 初始化结果为第一个点
	result := points[0]
	for _, point := range points[1:] {
		result = CalculatePoint(result, point, calculateMode, axis) // 使用CalculatePoint逐个比较
	}
	return result // 返回最终结果
}

// CanFormTriangle 判断三个点是否能构成三角形
// @param points 三个点的切片，必须包含三个点
// @return area 三角形的面积，如果能构成三角形则返回面积，
//
//	bool 表示是否能构成三角形，error 表示是否有错误
func CanFormTriangle(points []*gg.Point) (float64, bool, error) {
	maxPointLen := 3
	if len(points) < maxPointLen || len(points) > maxPointLen {
		return 0.0, false, fmt.Errorf("points length number must be %d", maxPointLen) // 如果点的数量不等于3，返回错误
	}

	// 计算三角形的面积
	// 公式的推导：
	// 给定三角形的三个顶点坐标 A(x1, y1)、B(x2, y2)、C(x3, y3)
	// 三角形的面积可以通过以下公式计算：
	// Area = 0.5 * | x1(y2 - y3) + x2(y3 - y1) + x3(y1 - y2) |
	// 其中，points[0] = A, points[1] = B, points[2] = C
	area := 0.5 * (points[0].X*(points[1].Y-points[2].Y) +
		points[1].X*(points[2].Y-points[0].Y) +
		points[2].X*(points[0].Y-points[1].Y))

	return area, area != 0, nil // 如果面积不为零，则可以构成三角形
}

// ResizeX 只缩放 x 坐标
// @param point 要缩放的点
// @param scaleFactor 缩放因子，>1 表示放大，<1 表示缩小
// @return 返回经过 x 坐标缩放后的新点
func ResizeX(point *gg.Point, scaleFactor float64) *gg.Point {
	return &gg.Point{
		X: point.X * scaleFactor, // 将 x 坐标乘以缩放因子
		Y: point.Y,               // y 坐标保持不变
	}
}

// ResizeY 只缩放 y 坐标
// @param point 要缩放的点
// @param scaleFactor 缩放因子，>1 表示放大，<1 表示缩小
// @return 返回经过 y 坐标缩放后的新点
func ResizeY(point *gg.Point, scaleFactor float64) *gg.Point {
	return &gg.Point{
		Y: point.Y * scaleFactor, // 将 y 坐标乘以缩放因子
		X: point.X,               // x 坐标保持不变
	}
}

// ResizePoint 单个坐标轴的等比例缩放
// @param point 要缩放的点
// @param resizeX x 坐标的缩放因子
// @param resizeY y 坐标的缩放因子
// @return 返回同时经过 x 和 y 坐标缩放后的新点
func ResizePoint(point *gg.Point, resizeX, resizeY float64) *gg.Point {
	return &gg.Point{
		X: ResizeX(point, resizeX).X, // 先对 x 坐标进行缩放
		Y: ResizeY(point, resizeY).Y, // 再对 y 坐标进行缩放
	}
}

// ResizePoints 批量缩放多个点，支持不同的 x 和 y 缩放因子
// @param points: 需要缩放的点的切片
// @param scaleX: x 轴的缩放因子
// @param scaleY: y 轴的缩放因子
// @return []*gg.Point: 返回缩放后的点的切片
func ResizePoints(points []*gg.Point, scaleX, scaleY float64) []*gg.Point {
	resizedPoints := make([]*gg.Point, len(points)) // 创建一个新的切片用于存储缩放后的点
	for i, point := range points {
		resizedPoints[i] = ResizePoint(point, scaleX, scaleY) // 使用缩放函数逐个缩放点
	}
	return resizedPoints // 返回缩放后的点的切片
}

// PointOperationStrategy 定义操作策略接口
type PointOperationStrategy interface {
	Apply(value, current float64) float64
}

// PointAddStrategy 实现加法操作
type PointAddStrategy struct{}

func (s *PointAddStrategy) Apply(value, current float64) float64 {
	return current + value
}

// PointSubtractStrategy 实现减法操作
type PointSubtractStrategy struct{}

func (s *PointSubtractStrategy) Apply(value, current float64) float64 {
	return current - value
}

// PointMultiplyStrategy 实现乘法操作
type PointMultiplyStrategy struct{}

func (s *PointMultiplyStrategy) Apply(value, current float64) float64 {
	return current * value
}

// PointDivideStrategy 实现除法操作
type PointDivideStrategy struct{}

func (s *PointDivideStrategy) Apply(value, current float64) float64 {
	return current / value
}

// ModifyPoint 根据指定的操作对点的某个坐标进行修改
// @param point 要修改的点
// @param value 要应用的值（缩放因子或偏移量）
// @param operation 计算模式
// @param isX true 表示修改 x 坐标，false 表示修改 y 坐标
// @return 返回经过操作后的新点
func ModifyPoint(point *gg.Point, value float64, operation CalculateFractionPointMode, isX bool) *gg.Point {
	var strategy PointOperationStrategy

	// 根据 operation 选择相应的策略
	switch operation {
	case Add:
		strategy = &PointAddStrategy{}
	case Subtract:
		strategy = &PointSubtractStrategy{}
	case Multiply:
		strategy = &PointMultiplyStrategy{}
	case Divide:
		strategy = &PointDivideStrategy{}
	default:
		strategy = &PointSubtractStrategy{} // 默认使用减法
	}

	// 选择要修改的坐标
	current := point.X
	if !isX {
		current = point.Y
	}

	// 应用策略
	newValue := strategy.Apply(value, current)

	// 返回新的点
	if isX {
		return &gg.Point{
			X: newValue,
			Y: point.Y, // y 坐标保持不变
		}
	}
	return &gg.Point{
		X: point.X, // x 坐标保持不变
		Y: newValue,
	}
}

// GetPointOperation 从可变参数中获取操作类型，默认为 Subtract
// @param operation 计算模式
// @return 返回操作类型
func GetPointOperation(operation ...CalculateFractionPointMode) CalculateFractionPointMode {
	if len(operation) > 0 {
		return operation[0]
	}
	return Subtract
}

// ResizePointOneselfX 根据指定的操作返回原始 y 坐标与缩放后的 x 坐标的结果
// @param point 要缩放的点
// @param scaleFactor 缩放因子，>1 表示放大，<1 表示缩小
// @param operation 计算模式
// @return 返回经过操作后的新点
func ResizePointOneselfX(point *gg.Point, scaleFactor float64, operation ...CalculateFractionPointMode) *gg.Point {
	operate := GetPointOperation(operation...)
	// 获取缩放后的 x 坐标
	resizedX := point.X * scaleFactor
	// 使用通用的 ModifyPoint 函数修改 x 坐标
	return ModifyPoint(point, resizedX, operate, true)
}

// ResizePointOneselfY 根据指定的操作返回原始 x 坐标与缩放后的 y 坐标的结果
// @param point 要缩放的点
// @param scaleFactor 缩放因子，>1 表示放大，<1 表示缩小
// @param operation 计算模式
// @return 返回经过操作后的新点
func ResizePointOneselfY(point *gg.Point, scaleFactor float64, operation ...CalculateFractionPointMode) *gg.Point {
	operate := GetPointOperation(operation...)
	// 获取缩放后的 y 坐标
	resizedY := point.Y * scaleFactor
	// 使用通用的 ModifyPoint 函数修改 y 坐标
	return ModifyPoint(point, resizedY, operate, false)
}

// OffsetPointX 根据指定的操作返回原始 y 坐标与偏移后的 x 坐标的结果
// @param point 要缩放的点
// @param offset 偏移量
// @param operation 计算模式
// @return 返回经过操作后的新点
func OffsetPointX(point *gg.Point, offset float64, operation ...CalculateFractionPointMode) *gg.Point {
	operate := GetPointOperation(operation...)
	// 使用通用的 ModifyPoint 函数修改 x 坐标
	return ModifyPoint(point, offset, operate, true)
}

// OffsetPointY 根据指定的操作返回原始 x 坐标与偏移后的 y 坐标的结果
// @param point 要缩放的点
// @param offset 偏移量
// @param operation 计算模式
// @return 返回经过操作后的新点
func OffsetPointY(point *gg.Point, offset float64, operation ...CalculateFractionPointMode) *gg.Point {
	operate := GetPointOperation(operation...)
	// 使用通用的 ModifyPoint 函数修改 y 坐标
	return ModifyPoint(point, offset, operate, false)
}

// ResizeUpTriangle 根据固定点放大三角形的两个顶点
// @params vertexA: 固定点的坐标
// @params vertexB: 需要放大的第一个点的坐标
// @params vertexC: 需要放大的第二个点的坐标
// @params resizeXy: 放大因子，值大于1表示放大
// @return: 新的 vertexB 和 vertexC 的坐标
func ResizeUpTriangle(vertexA, vertexB, vertexC *gg.Point, resizeXy float64) (*gg.Point, *gg.Point) {
	// 计算 vertexA 到 vertexB 和 vertexA 到 vertexC 的向量
	vectorAB := &gg.Point{X: vertexB.X - vertexA.X, Y: vertexB.Y - vertexA.Y}
	vectorAC := &gg.Point{X: vertexC.X - vertexA.X, Y: vertexC.Y - vertexA.Y}

	// 缩放向量
	resizedVectorAB := ResizePoint(vectorAB, resizeXy, resizeXy)
	resizedVectorAC := ResizePoint(vectorAC, resizeXy, resizeXy)

	// 计算放大后的 vertexB 和 vertexC 的新坐标
	newVertexB := &gg.Point{X: vertexA.X + resizedVectorAB.X, Y: vertexA.Y + resizedVectorAB.Y}
	newVertexC := &gg.Point{X: vertexA.X + resizedVectorAC.X, Y: vertexA.Y + resizedVectorAC.Y}

	return newVertexB, newVertexC
}

// ResizeDownTriangle 根据固定点缩小三角形的两个顶点
// @params vertexA: 固定点的坐标
// @params vertexB: 需要缩小的第一个点的坐标
// @params vertexC: 需要缩小的第二个点的坐标
// @params resizeXy: 缩小因子，值小于1表示缩小
// @return: 新的 vertexB 和 vertexC 的坐标
func ResizeDownTriangle(vertexA, vertexB, vertexC *gg.Point, resizeXy float64) (*gg.Point, *gg.Point) {
	// 确保缩小因子在0到1之间
	if resizeXy >= 1 || resizeXy <= 0 {
		return vertexB, vertexC // 如果因子不在范围内，返回原始坐标
	}

	// 计算 vertexA 到 vertexB 和 vertexA 到 vertexC 的向量
	vectorAB := &gg.Point{X: vertexB.X - vertexA.X, Y: vertexB.Y - vertexA.Y}
	vectorAC := &gg.Point{X: vertexC.X - vertexA.X, Y: vertexC.Y - vertexA.Y}

	// 缩放向量
	resizedVectorAB := ResizePoint(vectorAB, resizeXy, resizeXy)
	resizedVectorAC := ResizePoint(vectorAC, resizeXy, resizeXy)

	// 计算缩小后的 vertexB 和 vertexC 的新坐标
	newVertexB := &gg.Point{X: vertexA.X + resizedVectorAB.X, Y: vertexA.Y + resizedVectorAB.Y}
	newVertexC := &gg.Point{X: vertexA.X + resizedVectorAC.X, Y: vertexA.Y + resizedVectorAC.Y}

	return newVertexB, newVertexC
}

// ExtendLine 从起始点 p1 延长到 p2，返回延长线的终点坐标
// p1 是起始点，p2 是终止点，length 是延长的长度
func ExtendLine(p1, p2 *gg.Point, length float64) *gg.Point {
	// 计算线段的方向向量
	dx := p2.X - p1.X
	dy := p2.Y - p1.Y

	// 计算线段的长度
	lineLength := math.Sqrt(dx*dx + dy*dy)

	// 计算单位向量
	unitDx := dx / lineLength
	unitDy := dy / lineLength

	// 计算延长后的终点坐标
	newX := p1.X + unitDx*length
	newY := p1.Y + unitDy*length

	return &gg.Point{X: newX, Y: newY}
}

// AdjustValues 根据目标值调整起始和结束值
func AdjustValues(start, end, target float64) (float64, float64) {
	// 确保 start 小于 end
	if start > end {
		start, end = end, start
	}

	// 计算当前差值
	diff := end - start

	// 计算待补充的差值
	supplementalDifference := target - diff

	// 如果差值小于目标值，则进行调整
	if supplementalDifference > 0 {
		increment := supplementalDifference / 2
		start -= increment
		end += increment
	}

	return start, end
}
