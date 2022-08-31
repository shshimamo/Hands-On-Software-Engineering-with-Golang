package dependency

import "golang.org/x/xerrors"

type DepType int

const (
	DepTypeProject DepType = iota
	DepTypeResource
)

type API interface {
	ListDependencies(projectID string) ([]string, error)

	DependencyType(dependencyID string) (DepType, error)
}

type Collector struct {
	api API
}

func NewCollector(api API) *Collector {
	return &Collector{api: api}
}

func (c *Collector) AllDependencies(projectID string) ([]string, error) {
	ctx := newDepContext(projectID)
	for ctx.HasUncheckedDeps() {
		projectID = ctx.NextUncheckedDep()
		projectDeps, err := c.api.ListDependencies(projectID)
		if err != nil {
			return nil, xerrors.Errorf("unable to list dependencies for project %q: %w", projectID, err)
		}
		if err = c.scanProjectDependencies(ctx, projectDeps); err != nil {
			return nil, err
		}
	}
	return ctx.depList, nil
}

func (c *Collector) scanProjectDependencies(ctx *depCtx, depList []string) error {
	for _, depID := range depList {
		if ctx.AlreadyChecked(depID) {
			continue
		}
		ctx.AddToDepList(depID)
		depType, err := c.api.DependencyType(depID)
		if err != nil {
			return xerrors.Errorf("unable to get dependency type for id %q: %w", depID, err)
		}
		if depType == DepTypeProject {
			ctx.AddToUncheckedList(depID)
		}
	}
	return nil
}

type depCtx struct {
	depList   []string
	unchecked []string
	checked   map[string]struct{}
}

func newDepContext(projectID string) *depCtx {
	return &depCtx{
		unchecked: []string{projectID},
		checked:   make(map[string]struct{}),
	}
}

func (ctx *depCtx) HasUncheckedDeps() bool {
	return len(ctx.unchecked) != 0
}

func (ctx *depCtx) NextUncheckedDep() string {
	if len(ctx.unchecked) == 0 {
		return ""
	}

	next := ctx.unchecked[0]
	ctx.unchecked = ctx.unchecked[1:]
	return next
}

func (ctx *depCtx) AlreadyChecked(id string) bool {
	_, checked := ctx.checked[id]
	return checked
}

func (ctx *depCtx) AddToDepList(id string) {
	ctx.depList = append(ctx.depList, id)
	ctx.checked[id] = struct{}{}
}

func (ctx *depCtx) AddToUncheckedList(id string) {
	ctx.unchecked = append(ctx.unchecked, id)
}
