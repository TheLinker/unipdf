/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package creator

import (
	"fmt"

	"math"

	"github.com/unidoc/unidoc/common"
	"github.com/unidoc/unidoc/pdf/model/fonts"
)

type Chapter struct {
	number  int
	title   string
	heading *paragraph

	subchapters int

	contents []Drawable

	// Show chapter numbering
	showNumbering bool

	// Include in TOC.
	includeInTOC bool

	// Positioning: relative / absolute.
	positioning positioning

	// Absolute coordinates (when in absolute mode).
	xPos, yPos float64

	// Margins to be applied around the block when drawing on Page.
	margins margins

	// Chapter sizing is set to occupy available space.
	sizing Sizing

	// Reference to the creator's TOC.
	toc *TableOfContents
}

func (c *Creator) NewChapter(title string) *Chapter {
	chap := &Chapter{}

	c.chapters++
	chap.number = c.chapters
	chap.title = title

	chap.showNumbering = true
	chap.includeInTOC = true

	heading := fmt.Sprintf("%d. %s", c.chapters, title)
	p := NewParagraph(heading)
	p.SetFontSize(16)
	p.SetFont(fonts.NewFontHelvetica()) // bold?

	chap.heading = p
	chap.contents = []Drawable{}

	// Chapter sizing is fixed to occupy available size.
	chap.sizing = SizingOccupyAvailableSpace

	// Keep a reference for toc.
	chap.toc = c.toc

	return chap
}

// Set flag to indicate whether or not to show chapter numbers as part of title.
func (chap *Chapter) SetShowNumbering(show bool) {
	if show {
		heading := fmt.Sprintf("%d. %s", chap.number, chap.title)
		chap.heading.SetText(heading)
	} else {
		heading := fmt.Sprintf("%s", chap.title)
		chap.heading.SetText(heading)
	}
	chap.showNumbering = show
}

// Set flag to indicate whether or not to include in tOC.
func (chap *Chapter) SetIncludeInTOC(includeInTOC bool) {
	chap.includeInTOC = includeInTOC
}

func (chap *Chapter) GetHeading() *paragraph {
	return chap.heading
}

// Chapter sizing is fixed to occupy available space in the drawing context.
func (chap *Chapter) GetSizingMechanism() Sizing {
	return chap.sizing
}

// Chapter height is a sum of the content heights.
func (chap *Chapter) Height() float64 {
	h := float64(0)
	for _, d := range chap.contents {
		h += d.Height()
	}
	return h
}

// Chapter width is the maximum of the content widths.
func (chap *Chapter) Width() float64 {
	maxW := float64(0)
	for _, d := range chap.contents {
		maxW = math.Max(maxW, d.Width())
	}
	return maxW
}

// Set absolute coordinates.
func (chap *Chapter) SetPos(x, y float64) {
	chap.positioning = positionAbsolute
	chap.xPos = x
	chap.yPos = y
}

// Set chapter Margins.  Typically not needed as the Page Margins are used.
func (chap *Chapter) SetMargins(left, right, top, bottom float64) {
	chap.margins.left = left
	chap.margins.right = right
	chap.margins.top = top
	chap.margins.bottom = bottom
}

// Get chapter Margins: left, right, top, bottom.
func (chap *Chapter) GetMargins() (float64, float64, float64, float64) {
	return chap.margins.left, chap.margins.right, chap.margins.top, chap.margins.bottom
}

// Add a new drawable to the chapter.
func (chap *Chapter) Add(d Drawable) {
	if Drawable(chap) == d {
		common.Log.Debug("ERROR: Cannot add itself")
		return
	}

	switch d.(type) {
	case *Chapter:
		common.Log.Debug("Error: Cannot add chapter to a chapter")
	case *paragraph, *image, *Block, *subchapter:
		chap.contents = append(chap.contents, d)
	default:
		common.Log.Debug("Unsupported: %T", d)
	}
}

// Generate the Page blocks.  Multiple blocks are generated if the contents wrap over
// multiple pages.
func (chap *Chapter) GeneratePageBlocks(ctx DrawContext) ([]*Block, DrawContext, error) {
	blocks, ctx, err := chap.heading.GeneratePageBlocks(ctx)
	if err != nil {
		return blocks, ctx, err
	}
	if len(blocks) > 1 {
		ctx.Page++ // Did not fit, moved to new Page block.
	}

	if chap.includeInTOC {
		// Add to TOC.
		chap.toc.add(chap.title, chap.number, 0, ctx.Page)
	}

	for _, d := range chap.contents {
		newBlocks, c, err := d.GeneratePageBlocks(ctx)
		if err != nil {
			return blocks, ctx, err
		}
		if len(newBlocks) < 1 {
			continue
		}

		// The first block is always appended to the last..
		blocks[len(blocks)-1].mergeBlocks(newBlocks[0])
		blocks = append(blocks, newBlocks[1:]...)

		ctx = c
	}

	return blocks, ctx, nil
}
