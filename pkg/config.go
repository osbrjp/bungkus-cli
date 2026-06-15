package pkg

import (
	"errors"
	"maps"
	"slices"
	"time"
)

type (
	BaseFramework    string
	CSSFramework     string
	Formatter        string
	Linter           string
	ValidationLib    string
	FormLib          string
	QueryLib         string
	StateLib         string
	CMS              string
	PackageManager   string
	BaseGroup        string
	BaseIntegration  string
	TestingFramework string
	DeployTarget     string
	AuditTool        string
	CICDProvider     string
)

type AllDependencies struct {
	Dependencies
	DevDependencies
}

func (b BaseFramework) IsValid() bool {
	return globalRegistry != nil && globalRegistry.HasBase(string(b))
}

func (b BaseFramework) GetGroup(base string) (BaseGroup, error) {
	if globalRegistry == nil {
		return "", errors.New("unable to read registry")
	}

	entry := globalRegistry.GetBase(base)
	return BaseGroup(entry.Group), nil
}

func (b BaseFramework) GetDependencies() AllDependencies {
	if globalRegistry == nil {
		return AllDependencies{}
	}
	entry := globalRegistry.GetBase(string(b))
	if entry == nil {
		return AllDependencies{}
	}
	return AllDependencies{
		Dependencies:    entry.Packages.Dependencies,
		DevDependencies: entry.Packages.DevDependencies,
	}
}

func (b BaseFramework) IsAstro() bool {
	if globalRegistry == nil {
		return false
	}
	entry := globalRegistry.GetBase(string(b))
	return entry != nil && entry.Group == "astro"
}

func (b BaseFramework) IsNuxt() bool {
	if globalRegistry == nil {
		return false
	}
	entry := globalRegistry.GetBase(string(b))
	return entry != nil && entry.Group == "nuxt"
}

func (b BaseFramework) OutputDir() string {
	return "dist"
}

func (b BaseFramework) IsVite() bool {
	if globalRegistry == nil {
		return false
	}
	entry := globalRegistry.GetBase(string(b))
	return entry != nil && entry.Group == "vite"
}

func (b BaseFramework) GetIntegration() (BaseIntegration, error) {
	if globalRegistry != nil {
		return BaseIntegration(""), errors.New("unable to read registry")
	}

	entry := globalRegistry.GetBase(string(b))
	return BaseIntegration(entry.Integration), nil
}

// IsReactInt Check if integration is react
func (b BaseFramework) IsReactInt() bool {
	integration, err := b.GetIntegration()
	if err != nil {
		return false
	}

	return integration == "react"
}

// IsVueInt Check if integration is vue
func (b BaseFramework) IsVueInt() bool {
	integration, err := b.GetIntegration()
	if err != nil {
		return false
	}

	return integration == "react"
}

func (c CSSFramework) IsValid() bool {
	return globalRegistry != nil && globalRegistry.HasCSS(string(c))
}

func (c CSSFramework) GetDependencies() AllDependencies {
	if globalRegistry == nil {
		return AllDependencies{}
	}
	entry := globalRegistry.GetCSS(string(c))
	if entry == nil {
		return AllDependencies{}
	}
	return AllDependencies{
		Dependencies:    entry.Packages.Dependencies,
		DevDependencies: entry.Packages.DevDependencies,
	}
}

func (f Formatter) IsValid() bool {
	return globalRegistry != nil && globalRegistry.HasFormatter(string(f))
}

func (f Formatter) GetDependencies() AllDependencies {
	if globalRegistry == nil {
		return AllDependencies{}
	}
	entry := globalRegistry.GetFormatter(string(f))
	if entry == nil {
		return AllDependencies{}
	}
	return AllDependencies{
		Dependencies:    entry.Packages.Dependencies,
		DevDependencies: entry.Packages.DevDependencies,
	}
}

func (l Linter) IsValid() bool {
	return globalRegistry != nil && globalRegistry.HasLinter(string(l))
}

func (l Linter) GetDependencies() AllDependencies {
	if globalRegistry == nil {
		return AllDependencies{}
	}
	entry := globalRegistry.GetLinter(string(l))
	if entry == nil {
		return AllDependencies{}
	}
	return AllDependencies{
		Dependencies:    entry.Packages.Dependencies,
		DevDependencies: entry.Packages.DevDependencies,
	}
}

func (f FormLib) IsValid() bool {
	return globalRegistry != nil && globalRegistry.HasForm(string(f))
}

func (f FormLib) GetDependencies() AllDependencies {
	if globalRegistry == nil {
		return AllDependencies{}
	}
	entry := globalRegistry.GetForm(string(f))
	if entry == nil {
		return AllDependencies{}
	}
	return AllDependencies{
		Dependencies:    entry.Packages.Dependencies,
		DevDependencies: entry.Packages.DevDependencies,
	}
}

func (f FormLib) IsValidIntegration(base string) bool {
	if globalRegistry == nil {
		return false
	}
	form := globalRegistry.GetForm(string(f))
	if form == nil {
		return false
	}
	if len(form.RequiresIntegration) == 0 {
		return true
	}
	b := globalRegistry.GetBase(base)
	if b == nil {
		return false
	}
	effective := b.Integration
	if b.Group == "nuxt" {
		effective = "vue"
	}
	if effective == "" {
		return false
	}
	return slices.Contains(form.RequiresIntegration, effective)
}

func (t TestingFramework) IsValid() bool {
	return globalRegistry != nil && globalRegistry.HasTestingFramework(string(t))
}

func (t TestingFramework) GetDependencies() AllDependencies {
	if globalRegistry == nil {
		return AllDependencies{}
	}
	entry := globalRegistry.GetTestingFramework(string(t))
	if entry == nil {
		return AllDependencies{}
	}
	return AllDependencies{
		Dependencies:    entry.Packages.Dependencies,
		DevDependencies: entry.Packages.DevDependencies,
	}
}

func (a AuditTool) IsValid() bool {
	return globalRegistry != nil && globalRegistry.HasAudit(string(a))
}

func (a AuditTool) GetDependencies() AllDependencies {
	if globalRegistry == nil {
		return AllDependencies{}
	}
	entry := globalRegistry.GetAudit(string(a))
	if entry == nil {
		return AllDependencies{}
	}
	return AllDependencies{
		Dependencies:    entry.Packages.Dependencies,
		DevDependencies: entry.Packages.DevDependencies,
	}
}

func (v ValidationLib) IsValid() bool {
	return globalRegistry != nil && globalRegistry.HasValidation(string(v))
}

func (v ValidationLib) GetDependencies() AllDependencies {
	if globalRegistry == nil {
		return AllDependencies{}
	}
	entry := globalRegistry.GetValidation(string(v))
	if entry == nil {
		return AllDependencies{}
	}
	return AllDependencies{
		Dependencies:    entry.Packages.Dependencies,
		DevDependencies: entry.Packages.DevDependencies,
	}
}

func (q QueryLib) IsValid() bool {
	return globalRegistry != nil && globalRegistry.HasQuery(string(q))
}

func (q QueryLib) GetDependencies() AllDependencies {
	if globalRegistry == nil {
		return AllDependencies{}
	}
	entry := globalRegistry.GetQuery(string(q))
	if entry == nil {
		return AllDependencies{}
	}
	return AllDependencies{
		Dependencies:    entry.Packages.Dependencies,
		DevDependencies: entry.Packages.DevDependencies,
	}
}

func (q QueryLib) IsValidIntegration(base string) bool {
	if globalRegistry == nil {
		return false
	}
	query := globalRegistry.GetQuery(string(q))
	if query == nil {
		return false
	}
	if len(query.RequiresIntegration) == 0 {
		return true
	}
	b := globalRegistry.GetBase(base)
	if b == nil {
		return false
	}
	effective := b.Integration
	if b.Group == "nuxt" {
		effective = "vue"
	}
	if effective == "" {
		return false
	}
	return slices.Contains(query.RequiresIntegration, effective)
}

func (s StateLib) IsValid() bool {
	return globalRegistry != nil && globalRegistry.HasState(string(s))
}

func (s StateLib) GetDependencies() AllDependencies {
	if globalRegistry == nil {
		return AllDependencies{}
	}
	entry := globalRegistry.GetState(string(s))
	if entry == nil {
		return AllDependencies{}
	}
	return AllDependencies{
		Dependencies:    entry.Packages.Dependencies,
		DevDependencies: entry.Packages.DevDependencies,
	}
}

func (s StateLib) IsValidIntegration(base string) bool {
	if globalRegistry == nil {
		return false
	}
	state := globalRegistry.GetState(string(s))
	if state == nil {
		return false
	}
	if len(state.RequiresIntegration) == 0 {
		return true
	}
	b := globalRegistry.GetBase(base)
	if b == nil {
		return false
	}
	effective := b.Integration
	if b.Group == "nuxt" {
		effective = "vue"
	}
	if effective == "" {
		return false
	}
	return slices.Contains(state.RequiresIntegration, effective)
}

func (c CMS) IsValid() bool {
	return globalRegistry != nil && globalRegistry.HasCMS(string(c))
}

func (c CMS) GetDependencies() AllDependencies {
	if globalRegistry == nil {
		return AllDependencies{}
	}
	entry := globalRegistry.GetCMS(string(c))
	if entry == nil {
		return AllDependencies{}
	}
	return AllDependencies{
		Dependencies:    entry.Packages.Dependencies,
		DevDependencies: entry.Packages.DevDependencies,
	}
}

func (c CICDProvider) IsValid() bool {
	return globalRegistry != nil && globalRegistry.HasCICD(string(c))
}

func (c CICDProvider) GetDependencies() AllDependencies {
	if globalRegistry == nil {
		return AllDependencies{}
	}
	entry := globalRegistry.GetCICD(string(c))
	if entry == nil {
		return AllDependencies{}
	}
	return AllDependencies{
		Dependencies:    entry.Packages.Dependencies,
		DevDependencies: entry.Packages.DevDependencies,
	}
}

func (d DeployTarget) IsValid() bool {
	return globalRegistry != nil && globalRegistry.HasDeployment(string(d))
}

func (d DeployTarget) GetDependencies() AllDependencies {
	if globalRegistry == nil {
		return AllDependencies{}
	}
	entry := globalRegistry.GetDeployment(string(d))
	if entry == nil {
		return AllDependencies{}
	}
	return AllDependencies{
		Dependencies:    entry.Packages.Dependencies,
		DevDependencies: entry.Packages.DevDependencies,
	}
}

func (p PackageManager) IsValid() bool {
	return globalRegistry != nil && globalRegistry.HasPM(string(p))
}

func (p PackageManager) InstallCmd() string {
	if globalRegistry != nil {
		if entry := globalRegistry.GetPM(string(p)); entry != nil {
			return entry.InstallCmd
		}
	}
	return string(p) + " install"
}

func (p PackageManager) Exec() string {
	if globalRegistry != nil {
		if entry := globalRegistry.GetPM(string(p)); entry != nil {
			return entry.ExecCmd
		}
	}
	return "npx"
}

func (p PackageManager) RunCmd() string {
	if globalRegistry != nil {
		if entry := globalRegistry.GetPM(string(p)); entry != nil {
			return entry.RunCmd
		}
	}
	return string(p) + " dev"
}

type ProjectConfig struct {
	ProjectName string
	DestDir     string
	Date        string
	Site        string
	Base        BaseFramework
	CSS         CSSFramework
	Fmt         Formatter
	Linter      Linter
	Validation  ValidationLib
	Form        FormLib
	Query       QueryLib
	State       StateLib
	CMS         CMS
	Deployment  DeployTarget
	CICD        CICDProvider
	PM          PackageManager
	Test        TestingFramework
	Audit       AuditTool
}

// StackEntry is one row in the project's tech-stack table: the category
// (Framework, CSS, Linter, ...), the package name, and the resolved version.
type StackEntry struct {
	Tech    string
	Name    string
	Version string
}

// Stack returns a flattened, deterministic list of every package the selected
// options will install, grouped by tech category. Within each group, packages
// are sorted by name and dependencies/devDependencies are merged. Templates
// can render the README stacks table with a single range over the result.
func (c ProjectConfig) Stack() []StackEntry {
	var rows []StackEntry
	add := func(tech string, deps AllDependencies) {
		merged := make(map[string]string, len(deps.Dependencies)+len(deps.DevDependencies))
		maps.Copy(merged, deps.Dependencies)
		for n, v := range deps.DevDependencies {
			if _, ok := merged[n]; !ok {
				merged[n] = v
			}
		}
		names := make([]string, 0, len(merged))
		for n := range merged {
			names = append(names, n)
		}
		slices.Sort(names)
		for _, n := range names {
			rows = append(rows, StackEntry{Tech: tech, Name: n, Version: merged[n]})
		}
	}

	add("Framework", c.Base.GetDependencies())
	add("CSS", c.CSS.GetDependencies())
	add("Formatter", c.Fmt.GetDependencies())
	if string(c.Linter) != string(c.Fmt) {
		add("Linter", c.Linter.GetDependencies())
	}
	add("Validation", c.Validation.GetDependencies())
	add("Form", c.Form.GetDependencies())
	add("Query", c.Query.GetDependencies())
	add("State", c.State.GetDependencies())
	add("CMS", c.CMS.GetDependencies())
	add("Deployment", c.Deployment.GetDependencies())
	add("Testing", c.Test.GetDependencies())
	add("Audit", c.Audit.GetDependencies())
	return rows
}

func NewProjectConfig() ProjectConfig {
	return ProjectConfig{
		ProjectName: "my-app",
		Site:        "",
		Base:        "astro",
		CSS:         "vanilla",
		Fmt:         "biome",
		Linter:      "biome",
		Validation:  "none",
		Form:        "none",
		Query:       "none",
		State:       "none",
		CMS:         "none",
		PM:          "pnpm",
		Test:        "none",
		Deployment:  "none",
		CICD:        "none",
		Audit:       "none",
		Date:        time.Now().Format("2006-01-02"),
	}
}
