package timeslots

import (
	"sort"
)

// Map your struct to a Block.
type MapInFunc[In Period] func(In) *Block

// Map the Slot to your struct.
type MapOutFunc[Out any] func(*Slot) Out

// (Optional)Filter your struct in your condition.
type FilterFunc[Out any] func(Out) bool

// Options
type Options[Out any] struct {
	FilterFunc FilterFunc[Out]
}

// Whether the FilterFunc is set to Options
func (o *Options[Out]) IsSetFilter() bool {
	return o.FilterFunc != nil
}

// Option Func
type Option[Out any] func(*Options[Out])

// Run with filter option
func WithFilter[Out any](filter FilterFunc[Out]) Option[Out] {
	return func(opts *Options[Out]) {
		opts.FilterFunc = filter
	}
}

// Calculate available time slots (Slot). Provide the scheduled block (Block) and the target period (Span).
// Use this when passing and returning your struct.
func FindWithMapper[In Period, Out any](inputs []In, span *Span, mapin MapInFunc[In], mapout MapOutFunc[Out], opts ...Option[Out]) []Out {
	options := Options[Out]{
		FilterFunc: nil,
	}
	for _, opt := range opts {
		opt(&options)
	}

	if span == nil || !span.Remain() {
		return []Out{}
	}

	target := span.Clone()
	if len(inputs) == 0 {
		return []Out{mapout(target.ToSlot())}
	}

	sort.Slice(inputs, func(i, j int) bool {
		return inputs[i].Start().Before(inputs[j].Start())
	})

	j := 0
	slots := make([]Out, len(inputs)+1)
	for _, input := range inputs {
		block := mapin(input)
		
		if block.Contains(target) {
			target.Drop()
			break
		}

		if block.OverlapAtStart(target) {
			target.Shorten(block)
			continue
		}

		if block.IsContainedIn(target) {
			slot := mapout(createSlotFrom(target, block))
			target.Shorten(block)
			if options.IsSetFilter() && options.FilterFunc(slot) {
				continue
			}
			slots[j] = slot
			j++
			continue
		}

		if block.OverlapAtEnd(target) {
			slot := mapout(createSlotFrom(target, block))
			target.Drop()
			if options.IsSetFilter() && options.FilterFunc(slot) {
				break
			}
			slots[j] = slot
			j++
			break
		}
	}

	if !target.Remain() {
		return slots[:j]
	}
	slot := mapout(target.ToSlot())
	if options.IsSetFilter() && options.FilterFunc(slot) {
		return slots[:j]
	}
	slots[j] = slot
	j++
	return slots[:j]
}

// It returns a list of available time slots.
// Use this when passing and returning the pre-defined struct.
func Find(blocks []*Block, span *Span, opts ...Option[*Slot]) []*Slot {
	options := Options[*Slot]{
		FilterFunc: nil,
	}
	for _, opt := range opts {
		opt(&options)
	}

	if span == nil || !span.Remain() {
		return []*Slot{}
	}

	target := span.Clone()
	if len(blocks) == 0 {
		return []*Slot{target.ToSlot()}
	}

	sort.Slice(blocks, func(i, j int) bool {
		return blocks[i].Start().Before(blocks[j].Start())
	})

	j := 0
	slots := make([]*Slot, len(blocks)+1)
	for _, block := range blocks {
		
		if block.Contains(target) {
			target.Drop()
			break
		}

		if block.OverlapAtStart(target) {
			target.Shorten(block)
			continue
		}

		if block.IsContainedIn(target) {
			slot := createSlotFrom(target, block)
			target.Shorten(block)
			if options.IsSetFilter() && options.FilterFunc(slot) {
				continue
			}
			slots[j] = slot
			j++
			continue
		}

		if block.OverlapAtEnd(target) {
			slot := createSlotFrom(target, block)
			target.Drop()
			if options.IsSetFilter() && options.FilterFunc(slot) {
				break
			}
			slots[j] = slot
			j++
			break
		}
	}

	if !target.Remain() {
		return slots[:j]
	}
	slot := target.ToSlot()
	if options.IsSetFilter() && options.FilterFunc(slot) {
		return slots[:j]
	}
	slots[j] = slot
	j++
	return slots[:j]
}
