package testutil

import (
	"testing"

	"github.com/CedricHerzog/perfowl/internal/parser"
)

func TestProfileBuilder_Basic(t *testing.T) {
	profile := NewProfileBuilder().
		WithDuration(1000).
		Build()

	if profile == nil {
		t.Fatal("expected non-nil profile")
	}
	if profile.Duration() != 1000 {
		t.Errorf("Duration = %v, want 1000", profile.Duration())
	}
}

func TestProfileBuilder_WithMeta(t *testing.T) {
	meta := parser.Meta{
		Product:  "TestProduct",
		Platform: "TestPlatform",
	}
	profile := NewProfileBuilder().
		WithMeta(meta).
		Build()

	if profile.Meta.Product != "TestProduct" {
		t.Errorf("Product = %v, want TestProduct", profile.Meta.Product)
	}
}

func TestProfileBuilder_WithInterval(t *testing.T) {
	profile := NewProfileBuilder().
		WithInterval(2.0).
		Build()

	if profile.Meta.Interval != 2.0 {
		t.Errorf("Interval = %v, want 2.0", profile.Meta.Interval)
	}
}

func TestProfileBuilder_WithThread(t *testing.T) {
	thread := NewThreadBuilder("TestThread").Build()
	profile := NewProfileBuilder().
		WithThread(thread).
		Build()

	if len(profile.Threads) != 1 {
		t.Errorf("Threads length = %v, want 1", len(profile.Threads))
	}
	if profile.Threads[0].Name != "TestThread" {
		t.Errorf("Thread name = %v, want TestThread", profile.Threads[0].Name)
	}
}

func TestProfileBuilder_WithExtension(t *testing.T) {
	profile := NewProfileBuilder().
		WithExtension("ext@test", "Test Ext", "moz-extension://test/").
		Build()

	if profile.Meta.Extensions.Length != 1 {
		t.Errorf("Extensions length = %v, want 1", profile.Meta.Extensions.Length)
	}
}

func TestProfileBuilder_WithCategory(t *testing.T) {
	profile := NewProfileBuilder().
		WithCategory("TestCat", "blue").
		Build()

	// Default profile has 1 category, adding one makes 2
	found := false
	for _, cat := range profile.Meta.Categories {
		if cat.Name == "TestCat" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected to find TestCat category")
	}
}

func TestProfileBuilder_WithCategories(t *testing.T) {
	cats := []parser.Category{
		{Name: "Cat1", Color: "red"},
		{Name: "Cat2", Color: "blue"},
	}
	profile := NewProfileBuilder().
		WithCategories(cats).
		Build()

	if len(profile.Meta.Categories) != 2 {
		t.Errorf("Categories length = %v, want 2", len(profile.Meta.Categories))
	}
}

func TestProfileBuilder_WithSharedStringArray(t *testing.T) {
	profile := NewProfileBuilder().
		WithSharedStringArray([]string{"a", "b", "c"}).
		Build()

	if len(profile.Shared.StringArray) != 3 {
		t.Errorf("SharedStringArray length = %v, want 3", len(profile.Shared.StringArray))
	}
}

func TestThreadBuilder_Basic(t *testing.T) {
	thread := NewThreadBuilder("TestThread").Build()

	if thread.Name != "TestThread" {
		t.Errorf("Name = %v, want TestThread", thread.Name)
	}
}

func TestThreadBuilder_AsMainThread(t *testing.T) {
	thread := NewThreadBuilder("Main").AsMainThread().Build()

	if !thread.IsMainThread {
		t.Error("expected IsMainThread to be true")
	}
}

func TestThreadBuilder_WithProcessType(t *testing.T) {
	thread := NewThreadBuilder("Thread").WithProcessType("web").Build()

	if thread.ProcessType != "web" {
		t.Errorf("ProcessType = %v, want web", thread.ProcessType)
	}
}

func TestThreadBuilder_WithProcessName(t *testing.T) {
	thread := NewThreadBuilder("Thread").WithProcessName("Firefox").Build()

	if thread.ProcessName != "Firefox" {
		t.Errorf("ProcessName = %v, want Firefox", thread.ProcessName)
	}
}

func TestThreadBuilder_WithPID(t *testing.T) {
	thread := NewThreadBuilder("Thread").WithPID("12345").Build()

	if string(thread.PID) != "12345" {
		t.Errorf("PID = %v, want 12345", thread.PID)
	}
}

func TestThreadBuilder_WithTID(t *testing.T) {
	thread := NewThreadBuilder("Thread").WithTID("67890").Build()

	if string(thread.TID) != "67890" {
		t.Errorf("TID = %v, want 67890", thread.TID)
	}
}

func TestThreadBuilder_WithSamples(t *testing.T) {
	samples := NewSamplesBuilder().AddSample(0, 100).Build()
	thread := NewThreadBuilder("Thread").WithSamples(samples).Build()

	if thread.Samples.Length != 1 {
		t.Errorf("Samples length = %v, want 1", thread.Samples.Length)
	}
}

func TestThreadBuilder_WithMarkers(t *testing.T) {
	markers, _ := NewMarkerBuilder().AddGCMajor(0, 10).Build()
	thread := NewThreadBuilder("Thread").WithMarkers(markers).Build()

	if thread.Markers.Length != 1 {
		t.Errorf("Markers length = %v, want 1", thread.Markers.Length)
	}
}

func TestThreadBuilder_WithStackTable(t *testing.T) {
	st := NewStackTableBuilder().AddStack(0, 0, -1).Build()
	thread := NewThreadBuilder("Thread").WithStackTable(st).Build()

	if thread.StackTable.Length != 1 {
		t.Errorf("StackTable length = %v, want 1", thread.StackTable.Length)
	}
}

func TestThreadBuilder_WithFrameTable(t *testing.T) {
	ft := NewFrameTableBuilder().AddFrame(0, 0).Build()
	thread := NewThreadBuilder("Thread").WithFrameTable(ft).Build()

	if thread.FrameTable.Length != 1 {
		t.Errorf("FrameTable length = %v, want 1", thread.FrameTable.Length)
	}
}

func TestThreadBuilder_WithFuncTable(t *testing.T) {
	ft := NewFuncTableBuilder().AddFunc(0, true, -1).Build()
	thread := NewThreadBuilder("Thread").WithFuncTable(ft).Build()

	if thread.FuncTable.Length != 1 {
		t.Errorf("FuncTable length = %v, want 1", thread.FuncTable.Length)
	}
}

func TestThreadBuilder_WithResourceTable(t *testing.T) {
	rt := NewResourceTableBuilder().AddResource(0, 0, 0, 0).Build()
	thread := NewThreadBuilder("Thread").WithResourceTable(rt).Build()

	if thread.ResourceTable.Length != 1 {
		t.Errorf("ResourceTable length = %v, want 1", thread.ResourceTable.Length)
	}
}

func TestThreadBuilder_WithStringArray(t *testing.T) {
	thread := NewThreadBuilder("Thread").WithStringArray([]string{"a", "b"}).Build()

	if len(thread.StringArray) != 2 {
		t.Errorf("StringArray length = %v, want 2", len(thread.StringArray))
	}
}

func TestMarkerBuilder_Basic(t *testing.T) {
	mb := NewMarkerBuilder()
	mb.AddGCMajor(0, 10)
	mb.AddGCMinor(20, 5)
	mb.AddGCSlice(30, 3)
	mb.AddLongTask(50, 100)
	markers, strings := mb.Build()

	if markers.Length != 4 {
		t.Errorf("Markers length = %v, want 4", markers.Length)
	}
	if len(strings) == 0 {
		t.Error("expected non-empty string array")
	}
}

func TestMarkerBuilder_DOMEvent(t *testing.T) {
	mb := NewMarkerBuilder()
	mb.AddDOMEvent("click", 0)
	mb.AddDOMEventWithDuration("input", 10, 5)
	markers, _ := mb.Build()

	if markers.Length != 2 {
		t.Errorf("Markers length = %v, want 2", markers.Length)
	}
}

func TestMarkerBuilder_IPC(t *testing.T) {
	mb := NewMarkerBuilder()
	mb.AddIPC(0, 10, true)
	mb.AddIPC(20, 5, false)
	markers, _ := mb.Build()

	if markers.Length != 2 {
		t.Errorf("Markers length = %v, want 2", markers.Length)
	}
}

func TestMarkerBuilder_Layout(t *testing.T) {
	mb := NewMarkerBuilder()
	mb.AddLayout(0, 10)
	mb.AddStyles(20, 5)
	markers, _ := mb.Build()

	if markers.Length != 2 {
		t.Errorf("Markers length = %v, want 2", markers.Length)
	}
}

func TestMarkerBuilder_Network(t *testing.T) {
	mb := NewMarkerBuilder()
	mb.AddNetwork("https://example.com", 0, 100)
	mb.AddChannelMarker("https://test.com", 50, 200)
	mb.AddHostResolver("example.com", 10, 30)
	markers, _ := mb.Build()

	if markers.Length != 3 {
		t.Errorf("Markers length = %v, want 3", markers.Length)
	}
}

func TestMarkerBuilder_UserTiming(t *testing.T) {
	mb := NewMarkerBuilder()
	mb.AddUserTiming("mark1", 0)
	mb.AddUserTimingWithDuration("measure1", 10, 50)
	markers, _ := mb.Build()

	if markers.Length != 2 {
		t.Errorf("Markers length = %v, want 2", markers.Length)
	}
}

func TestMarkerBuilder_Misc(t *testing.T) {
	mb := NewMarkerBuilder()
	mb.AddAwake(0)
	mb.AddJSActorMessage("TestActor", 10, 5)
	mb.AddPaint(20, 10)
	mb.AddText("test text", 30)
	markers, _ := mb.Build()

	if markers.Length != 4 {
		t.Errorf("Markers length = %v, want 4", markers.Length)
	}
}

func TestMarkerBuilder_Chrome(t *testing.T) {
	mb := NewMarkerBuilder()
	mb.AddEventDispatch("click", 0, 5)
	mb.AddUpdateLayoutTree(10, 15)
	mb.AddV8GCScavenger(30, 5)
	mb.AddV8GCMarkCompactor(40, 20)
	mb.AddMajorGC(60, 10)
	mb.AddMinorGC(80, 3)
	markers, _ := mb.Build()

	if markers.Length != 6 {
		t.Errorf("Markers length = %v, want 6", markers.Length)
	}
}

func TestMarkerBuilder_Custom(t *testing.T) {
	mb := NewMarkerBuilder()
	mb.AddCustom("CustomMarker", 0, 100, 50, map[string]interface{}{
		"custom_field": "value",
	})
	markers, _ := mb.Build()

	if markers.Length != 1 {
		t.Errorf("Markers length = %v, want 1", markers.Length)
	}
}

func TestMarkerBuilder_BuildMarkers(t *testing.T) {
	mb := NewMarkerBuilder()
	mb.AddGCMajor(0, 10)
	markers := mb.BuildMarkers()

	if markers.Length != 1 {
		t.Errorf("Markers length = %v, want 1", markers.Length)
	}
}

func TestMarkerBuilder_BuildForThread(t *testing.T) {
	mb := NewMarkerBuilder()
	mb.AddGCMajor(0, 10)

	tb := NewThreadBuilder("TestThread")
	mb.BuildForThread(tb)
	thread := tb.Build()

	if thread.Markers.Length != 1 {
		t.Errorf("Markers length = %v, want 1", thread.Markers.Length)
	}
}

func TestSamplesBuilder_Basic(t *testing.T) {
	sb := NewSamplesBuilder()
	sb.AddSample(0, 0)
	sb.AddSample(0, 10)
	samples := sb.Build()

	if samples.Length != 2 {
		t.Errorf("Samples length = %v, want 2", samples.Length)
	}
}

func TestSamplesBuilder_WithCPUDelta(t *testing.T) {
	sb := NewSamplesBuilder()
	sb.AddSampleWithCPUDelta(0, 0, 1000)
	samples := sb.Build()

	if samples.Length != 1 {
		t.Errorf("Samples length = %v, want 1", samples.Length)
	}
	if samples.ThreadCPUDelta[0] != 1000 {
		t.Errorf("ThreadCPUDelta = %v, want 1000", samples.ThreadCPUDelta[0])
	}
}

func TestSamplesBuilder_WithWeight(t *testing.T) {
	sb := NewSamplesBuilder()
	sb.AddSampleWithWeight(0, 0, 5)
	samples := sb.Build()

	if samples.Length != 1 {
		t.Errorf("Samples length = %v, want 1", samples.Length)
	}
	if samples.Weight[0] != 5 {
		t.Errorf("Weight = %v, want 5", samples.Weight[0])
	}
}

func TestSamplesBuilder_WithWeightType(t *testing.T) {
	sb := NewSamplesBuilder()
	sb.WithWeightType("samples")
	samples := sb.Build()

	if samples.WeightType != "samples" {
		t.Errorf("WeightType = %v, want samples", samples.WeightType)
	}
}

func TestStackTableBuilder(t *testing.T) {
	stb := NewStackTableBuilder()
	stb.AddStack(0, 0, -1)
	stb.AddStack(1, 0, 0)
	st := stb.Build()

	if st.Length != 2 {
		t.Errorf("StackTable length = %v, want 2", st.Length)
	}
}

func TestFrameTableBuilder(t *testing.T) {
	ftb := NewFrameTableBuilder()
	ftb.AddFrame(0, 0)
	ftb.AddFrame(1, 1)
	ft := ftb.Build()

	if ft.Length != 2 {
		t.Errorf("FrameTable length = %v, want 2", ft.Length)
	}
}

func TestFuncTableBuilder(t *testing.T) {
	fnb := NewFuncTableBuilder()
	fnb.AddFunc(0, true, -1)
	fnb.AddFuncWithFile(1, false, 0, 2, 100)
	ft := fnb.Build()

	if ft.Length != 2 {
		t.Errorf("FuncTable length = %v, want 2", ft.Length)
	}
}

func TestResourceTableBuilder(t *testing.T) {
	rtb := NewResourceTableBuilder()
	rtb.AddResource(0, 1, 2, 3)
	rt := rtb.Build()

	if rt.Length != 1 {
		t.Errorf("ResourceTable length = %v, want 1", rt.Length)
	}
}

func TestDefaultCategories(t *testing.T) {
	cats := DefaultCategories()

	if len(cats) == 0 {
		t.Error("expected non-empty categories")
	}

	// Check a known category exists
	found := false
	for _, cat := range cats {
		if cat.Name == "JavaScript" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected JavaScript category")
	}
}
