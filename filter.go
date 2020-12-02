package filter

import (
	"context"
	"fmt"
	"sort"

	"github.com/pkg/errors"

	"github.com/techxmind/filter/core"
)

var (
	_ = fmt.Println
)

// Filter interface
type Filter interface {
	Name() string
	Run(ctx context.Context, data interface{}) bool
}

//singleFilter contains single filter
type singleFilter struct {
	name      string
	condition core.Condition
	executor  core.Executor
}

func (f *singleFilter) Name() string { return f.name }
func (f *singleFilter) Run(pctx context.Context, data interface{}) bool {
	ctx := core.WithData(pctx, data)
	trace := ctx.Trace()

	if trace != nil {
		trace.Enter("COND")
	}

	ok := f.condition.Success(ctx)

	if trace != nil {
		trace.Leave("COND").Log("RET", ok)
	}

	if !ok {
		return false
	}

	if f.executor != nil {
		if trace != nil {
			trace.Enter("EXEC")
		}

		f.executor.Execute(ctx, data)

		if trace != nil {
			trace.Leave("EXEC")
		}
	}
	return true
}

// FilterGroup contains multiple filters
type FilterGroup struct {
	name    string
	filters []Filter
	// when shortMode = true, Run method will return immediately when find the first filter that return true.
	shortMode bool

	// when enableRank = true, run filters with rank order and usually is running in shortMode.
	enableRank   bool
	ranks        []*rank
	rankBoundary []*rankBoundary
}

func NewFilterGroup(options ...Option) *FilterGroup {
	opts := getFilterOpts(options)

	return &FilterGroup{
		name:         opts.name,
		filters:      make([]Filter, 0),
		shortMode:    opts.shortMode,
		enableRank:   opts.enableRank,
		ranks:        make([]*rank, 0),
		rankBoundary: make([]*rankBoundary, 0),
	}
}

type rank struct {
	idx      int   // index of filter
	weight   int64 // weight of filter
	priority int64 // priority of filter
}

func (r rank) Weight() int64 {
	return r.weight
}

type rankBoundary struct {
	boundary    int   // priority boundary of rank
	totalWeight int64 // total weight of priority
}

func (f *FilterGroup) Name() string { return f.name }

func (f *FilterGroup) Run(pctx context.Context, data interface{}) (succ bool) {
	ctx := core.WithData(pctx, data)
	trace := ctx.Trace()

	idxes := make([]int, len(f.filters))
	for i, _ := range f.filters {
		idxes[i] = i
	}

	if trace != nil {
		trace.Enter("FILTER " + f.Name())
	}

	if f.enableRank {
		// sort filter by priority desc
		for i, rank := range f.ranks {
			idxes[i] = rank.idx
		}
		// shuffle each priority group items by probability(weight)
		lastIdx := 0
		for _, b := range f.rankBoundary {
			itemCount := b.boundary - lastIdx
			if itemCount <= 1 {
				lastIdx = b.boundary
				continue
			}
			totalWeight := b.totalWeight
			items := make([]Weighter, itemCount)
			for i := lastIdx; i < b.boundary; i++ {
				items[i] = f.ranks[i]
			}

			for i := 0; i < len(items); i++ {
				idx := i + PickIndexByWeight(items[i:], totalWeight)
				items[idx], items[i] = items[i], items[idx]
				totalWeight -= items[i].Weight()
			}
			partIdxes := make([]int, len(items))
			for i, item := range items {
				partIdxes[i] = item.(*rank).idx
			}
			copy(idxes[lastIdx:b.boundary], partIdxes)
			lastIdx = b.boundary
		}

		if trace != nil {
			trace.Log("RANK ", idxes)
		}
	}

	for _, idx := range idxes {
		filter := f.filters[idx]
		if trace != nil {
			trace.Enter("FILTER " + filter.Name())
		}

		isucc := filter.Run(ctx, data)

		if trace != nil {
			trace.Leave("FILTER "+filter.Name()).Log("RET", isucc)
		}
		if isucc {
			succ = isucc
			if f.shortMode {
				if trace != nil {
					trace.Leave("END "+filter.Name()).Log("RET", succ)
				}
				return
			}
		}
	}

	if trace != nil {
		trace.Leave("END "+f.Name()).Log("RET", succ)
	}
	return
}

func (f *FilterGroup) Add(filter Filter, options ...Option) {
	opts := getFilterOpts(options)

	f.filters = append(f.filters, filter)

	if !f.enableRank {
		return
	}

	f.ranks = append(f.ranks, &rank{
		idx:      len(f.filters) - 1,
		weight:   opts.weight,
		priority: opts.priority,
	})

	// sort by priority desc
	sort.Slice(f.ranks, func(i, j int) bool {
		return f.ranks[i].priority > f.ranks[j].priority
	})

	// group by priority, set group boundary and total weight for later probability calculation
	f.rankBoundary = f.rankBoundary[:0]
	lastPriority := int64(-1)
	totalWeight := int64(0)
	for i, rank := range f.ranks {
		if i != 0 && rank.priority != lastPriority {
			f.rankBoundary = append(f.rankBoundary, &rankBoundary{
				boundary:    i,
				totalWeight: totalWeight,
			})
			totalWeight = 0
		}
		totalWeight += rank.weight
		lastPriority = rank.priority
	}
	f.rankBoundary = append(f.rankBoundary, &rankBoundary{
		boundary:    len(f.ranks),
		totalWeight: totalWeight,
	})
}

type Options struct {
	weight     int64
	priority   int64
	shortMode  bool
	enableRank bool
	name       string
	namePrefix string
}

type Option interface {
	apply(*Options)
}

// Weight Option
type Weight uint64

func (o Weight) apply(opts *Options) {
	opts.weight = int64(o)
}

// Priority Option
type Priority uint64

func (o Priority) apply(opts *Options) {
	opts.priority = int64(o)
}

// ShortMode Option
type ShortMode bool

func (o ShortMode) apply(opts *Options) {
	opts.shortMode = bool(o)
}

// EnableRank Option
type EnableRank bool

func (o EnableRank) apply(opts *Options) {
	opts.enableRank = bool(o)
	// auto set short mode
	opts.shortMode = true
}

// Name Option
type Name string

func (o Name) apply(opts *Options) {
	opts.name = string(o)
}

// NamePrefix Option
type NamePrefix string

func (o NamePrefix) apply(opts *Options) {
	opts.namePrefix = string(o)
}

// getFilterOpts return *Options
func getFilterOpts(opts []Option) *Options {
	o := &Options{}

	for _, opt := range opts {
		opt.apply(o)
	}

	return o
}

// New build filter with specified data struct.
//  items:
//    single filter:
//     [
//       "$filter-name"  // filter name, first item, optional
//       ["$var-name", "$op", "$op-value"],  // condition
//       ["$var-name", "$op", "$op-value"],  // condition
//       ["$data-key", "$assign", "$assign-value"] // executor, last item
//     ]
//
//    filter group:
//    [
//      // filter
//      [
//       "$filter-name"  // filter name, first item, optional
//       ["$var-name", "$op", "$op-value"],  // condition
//       ["$var-name", "$op", "$op-value"],  // condition
//       ["$data-key", "$assign", "$assign-value"] // executor, last item
//     ],
//     // filter
//     [
//       "$filter-name"  // filter name, first item, optional
//       ["$var-name", "$op", "$op-value"],  // condition
//       ["$var-name", "$op", "$op-value"],  // condition
//       ["$data-key", "$assign", "$assign-value"] // executor, last item
//     ]
//    ]
//
//  options:
//    ShortMode(true)     // enable short mode, only active in group filter
//    EnableRank(true)    // enable rank mode, and set short mode only active in group filter
//    Name("filter-name") // specify filter name
//
func New(items []interface{}, options ...Option) (Filter, error) {
	if len(items) == 0 {
		return nil, errors.New("Empty filter")
	}

	if !core.IsArray(items[0]) {
		return nil, errors.New("Filter data error,first element is not array")
	}

	item := core.ToArray(items[0])
	if len(item) == 0 {
		return nil, errors.New("Filter data error,first element is empty array")
	}

	// single filter
	if !core.IsArray(item[0]) {
		return buildFilter(items, options...)
	}

	group := NewFilterGroup(options...)

	if group.name == "" {
		group.name = generateFilterName(items)
	}

	for _, item := range items {
		if !core.IsArray(item) {
			return nil, errors.New("Filter group data error,element must be array")
		}
		if filter, err := buildFilter(core.ToArray(item), NamePrefix(group.name+".")); err != nil {
			return nil, err
		} else {
			group.Add(filter)
		}
	}

	return group, nil
}

// buildFilter build filter with data.
// [
//   "$filter-name"  // filter name, first item, optional
//   ["$var-name", "$op", "$op-value"],  // condition
//   ["$var-name", "$op", "$op-value"],  // condition
//   ["$data-key", "$assign", "$assign-value"] // executor, last item
// ]
//
func buildFilter(data []interface{}, options ...Option) (Filter, error) {
	if len(data) == 0 {
		return nil, errors.New("Filter struct is empty")
	}

	opts := getFilterOpts(options)

	// filter name
	name, ok := data[0].(string)
	if ok {
		data = data[1:]
	} else {
		if opts.name != "" {
			name = opts.name
		} else {
			name = generateFilterName(data)
		}
	}

	if opts.namePrefix != "" {
		name = opts.namePrefix + name
	}

	if len(data) < 2 {
		return nil, errors.New("Filter struct must contain conditions and assigment")
	}

	filter := &singleFilter{
		name: name,
	}

	if condition, err := core.NewCondition(data[:len(data)-1], core.LOGIC_ALL); err != nil {
		return nil, errors.Wrap(err, "condition")
	} else {
		filter.condition = condition
	}

	if data[len(data)-1] != nil {
		if executor, err := core.NewExecutor(data[len(data)-1:]); err != nil {
			return nil, errors.Wrap(err, "executor")
		} else {
			filter.executor = executor
		}
	}

	return filter, nil
}
