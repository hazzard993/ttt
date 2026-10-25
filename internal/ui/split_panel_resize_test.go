package ui

import (
	"testing"

	"github.com/eugenioenko/ttt/internal/term"

	"github.com/gdamore/tcell/v3"
)

func newResizeTestSplit() (*SplitPanelWidget, *mockWidget, *[]int) {
	right := &mockWidget{}
	split := NewSplitPanelWidget()
	split.Left = &mockWidget{}
	split.Right = right
	split.DividerPos = 20
	split.ShowLeft = true
	split.SetRect(Rect{W: 60, H: 10})
	var resizes []int
	split.OnResize = func(w int) { resizes = append(resizes, w) }
	return split, right, &resizes
}

func TestSplitPanelStationaryClickOnRightEdgeReachesRight(t *testing.T) {
	split, right, resizes := newResizeTestSplit()
	x := split.DividerScreenX() + 1

	split.HandleEvent(tcell.NewEventMouse(x, 3, tcell.Button1, 0))
	split.HandleEvent(tcell.NewEventMouse(x, 3, tcell.ButtonNone, 0))

	if right.eventCount != 2 {
		t.Fatalf("right panel events = %d, want replayed press and release", right.eventCount)
	}
	if len(*resizes) != 0 {
		t.Fatalf("stationary click resized: %v", *resizes)
	}
}

func TestSplitPanelVerticalMoveOnRightEdgeIsNotAClick(t *testing.T) {
	split, right, resizes := newResizeTestSplit()
	x := split.DividerScreenX() + 1

	split.HandleEvent(tcell.NewEventMouse(x, 3, tcell.Button1, 0))
	split.HandleEvent(tcell.NewEventMouse(x, 6, tcell.Button1, 0))
	split.HandleEvent(tcell.NewEventMouse(x, 6, tcell.ButtonNone, 0))

	if right.eventCount != 0 {
		t.Fatalf("right panel received %d events from a vertical drag", right.eventCount)
	}
	if len(*resizes) != 0 {
		t.Fatalf("vertical-only drag resized: %v", *resizes)
	}
	if split.Dragging() || split.OwnsPointerCapture() {
		t.Fatal("drag state survived release")
	}
}

func TestSplitPanelHorizontalDragFromRightEdgeResizes(t *testing.T) {
	split, right, resizes := newResizeTestSplit()
	x := split.DividerScreenX() + 1

	split.HandleEvent(tcell.NewEventMouse(x, 3, tcell.Button1, 0))
	split.HandleEvent(tcell.NewEventMouse(x+4, 3, tcell.Button1, 0))
	split.HandleEvent(tcell.NewEventMouse(x+4, 3, tcell.ButtonNone, 0))

	if right.eventCount != 0 {
		t.Fatalf("right panel received %d events from a resize drag", right.eventCount)
	}
	if len(*resizes) == 0 || (*resizes)[len(*resizes)-1] != x+4-1 {
		t.Fatalf("resizes = %v, want final width %d", *resizes, x+3)
	}
}

func TestSidebarRightLayoutAndInteraction(t *testing.T) {
	split, content, resizes := newResizeTestSplit()
	sidebar := &mockWidget{renderChar: 'S'}
	split.Left = sidebar
	split.SidebarPosition = SidebarRight
	content.renderChar = 'C'
	cells := makeGrid(60, 10)
	split.Render(NewRenderSurface(cells, Rect{W: 60, H: 10}))
	if sidebar.GetRect() != (Rect{X: 39, Y: 1, W: 20, H: 8}) || content.GetRect() != (Rect{X: 1, W: 37, H: 9}) {
		t.Fatalf("sidebar %v, content %v", sidebar.GetRect(), content.GetRect())
	}
	if cells[3][40].Ch != 'S' || cells[3][20].Ch != 'C' || cells[0][38].Ch != term.SingleBorderSet().TopLeft {
		t.Fatal("right sidebar rendered in wrong position")
	}
	split.HandleEvent(tcell.NewEventMouse(40, 3, tcell.Button1, 0))
	split.HandleEvent(tcell.NewEventMouse(40, 3, tcell.ButtonNone, 0))
	if sidebar.eventCount != 2 || content.eventCount != 0 {
		t.Fatal("sidebar click reached wrong child")
	}
	x := split.DividerScreenX()
	split.HandleEvent(tcell.NewEventMouse(x, 3, tcell.Button1, 0))
	split.HandleEvent(tcell.NewEventMouse(x-4, 3, tcell.Button1, 0))
	split.HandleEvent(tcell.NewEventMouse(x-4, 3, tcell.ButtonNone, 0))
	if len(*resizes) != 1 || (*resizes)[0] != 24 {
		t.Fatalf("right sidebar drag widths = %v, want [24]", *resizes)
	}
	split.ShowLeft = false
	if !split.OverDivider(59, 3) || split.OverDivider(0, 3) {
		t.Fatal("hidden sidebar resize edge is wrong")
	}
	split.Render(NewRenderSurface(cells, Rect{W: 60, H: 10}))
	if content.GetRect() != (Rect{X: 1, W: 58, H: 9}) {
		t.Fatalf("hidden sidebar content rect = %v", content.GetRect())
	}
}
