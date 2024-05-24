package view

import (
	"fmt"
	"sort"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// NewApp starts the explorer pointed at the given chain URL and ID
func NewApp(chainURL string, chainID uint64) (*App, error) {
	app := &App{
		Application: tview.NewApplication(),
		chainURL:    chainURL,
		chainID:     chainID,
		fullTopView: tview.NewFlex().SetDirection(tview.FlexRow),
		Context:     NewContextView(chainURL, chainID),
		Control:     NewControlView(),
		Main:        tview.NewFlex(),
	}
	app.Context.SetChangedFunc(func() {
		app.Draw()
	})
	app.Control.SetChangedFunc(func() {
		app.Draw()
	})
	app.Control.defaultControls = ControlMapping{
		NormalControls: NormalKeyControls{
			'/': Control{
				Key:         "/",
				Description: "Search",
				Order:       0,
				Fn:          app.Search,
			},
			'h': Control{
				Key:         "h",
				Description: "Home",
				Order:       1,
				Fn: func() {
					client, err := ethclient.Dial(chainURL)
					if err != nil {
						app.UpdateContext(fmt.Sprintf("[red]Error connecting to chain at %s: %s[-]", chainURL, err.Error()))
						return
					}
					app.ShowChainSummary(client)
				},
			},
		},
		SpecialControls: SpecialKeyControls{
			tcell.KeyCtrlC: Control{
				Key:         "Ctrl+C",
				Description: "Quit",
				Order:       2,
				Fn: func() {
					app.Stop()
				},
			},
		},
	}
	app.UpdateControls(nil)

	// Build formatting containers
	app.contextAndControlView = tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(app.Context, 0, 1, false).
		AddItem(app.Control, 0, 1, false)
	app.fullTopView.AddItem(app.contextAndControlView, 0, 1, false)

	app.fullView = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(app.fullTopView, 0, 1, false).
		AddItem(app.Main, 0, 4, true)

	app.SetRoot(app.fullView, true)

	client, err := ethclient.Dial(chainURL)
	if err != nil {
		return nil, fmt.Errorf("Error connecting to chain at %s: %s", chainURL, err.Error())
	}

	app.ShowChainSummary(client)
	return app, nil
}

type App struct {
	*tview.Application

	chainURL string
	chainID  uint64

	// fullView wraps the entire application
	fullView *tview.Flex
	// contextAndControlView holds the context and control views
	contextAndControlView *tview.Flex
	// fullTopView holds the contextAndControlView and any other top-level views (search)
	fullTopView *tview.Flex

	Context *Context
	Control *Controls
	Main    *tview.Flex
}

// Context is the view that displays the current context of the application, usually the chain URL and ID, but also errors and loading messages
type Context struct {
	*tview.TextView

	permanentText string
}

func NewContextView(chainURL string, chainID uint64) *Context {
	permanentText := fmt.Sprintf("URL: %s\nID: %d", chainURL, chainID)
	view := tview.NewTextView().
		SetText(permanentText).
		SetDynamicColors(true).
		SetRegions(true).
		SetWordWrap(true).
		SetWrap(true)
	view.SetBorder(true)
	return &Context{
		TextView:      view,
		permanentText: permanentText,
	}
}

// Update updates the context view with the given text
func (app *App) UpdateContext(text string) {
	app.Context.Clear()
	app.Context.SetText(fmt.Sprintf("%s\n%s", app.Context.permanentText, text))
}

// Add adds the given text to the context view without clearing the existing text
func (app *App) AddContext(text string) {
	app.Context.SetText(fmt.Sprintf("%s\n%s", app.Context.GetText(false), text))
}

// ControlMapping is a mapping of controls to their key and description
type ControlMapping struct {
	NormalControls  NormalKeyControls
	SpecialControls SpecialKeyControls
}

// NormalKeyControls is a mapping of a "normal" key press (rune) to a Control
type NormalKeyControls map[rune]Control

// SpecialKeyControls is a mapping of a "special" key press (e.g. CTRL+C) to a Control
type SpecialKeyControls map[tcell.Key]Control

// Control is a key press that triggers a function
type Control struct {
	Key         string
	Description string
	// Order is used to sort controls when displayed to the user. Default controls are always first
	Order uint
	Fn    func()
}

// Controls is the view that displays the current controls available to the user
type Controls struct {
	*tview.TextView

	// defaultControls are the controls that are always available
	defaultControls ControlMapping
}

func NewControlView() *Controls {
	view := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWordWrap(true).
		SetWrap(true)
	view.SetBorder(true)
	return &Controls{
		TextView: view,
	}
}

// UpdateControls updates what controls the app can use and displays them in the control view
func (app *App) UpdateControls(controls *ControlMapping) {
	if controls == nil {
		controls = &ControlMapping{}
	}
	if controls.NormalControls == nil {
		controls.NormalControls = NormalKeyControls{}
	}
	if controls.SpecialControls == nil {
		controls.SpecialControls = SpecialKeyControls{}
	}

	var sortedControls []Control
	for _, control := range controls.NormalControls {
		sortedControls = append(sortedControls, control)
	}
	for _, control := range controls.SpecialControls {
		sortedControls = append(sortedControls, control)
	}
	// Sort controls by order
	sort.Slice(sortedControls, func(i, j int) bool {
		return sortedControls[i].Order < sortedControls[j].Order
	})

	// Merge default controls with new controls
	for key, defaultControl := range app.Control.defaultControls.SpecialControls {
		sortedControls = append([]Control{defaultControl}, sortedControls...)
		controls.SpecialControls[key] = defaultControl
	}
	for key, defaultControl := range app.Control.defaultControls.NormalControls {
		sortedControls = append([]Control{defaultControl}, sortedControls...)
		controls.NormalControls[key] = defaultControl
	}

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyRune:
			if control, exists := controls.NormalControls[event.Rune()]; exists {
				if control.Fn == nil {
					app.UpdateContext(fmt.Sprintf("[red]No function for control %c[-]", event.Rune()))
					return nil
				}
				control.Fn()
			}
		default:
			if control, exists := controls.SpecialControls[event.Key()]; exists {
				if control.Fn == nil {
					return nil
				}
				control.Fn()
			}
		}
		return event
	})

	controlsText := ""
	for _, control := range sortedControls {
		controlsText = fmt.Sprintf("%s[blue]%s[-]: %s\n", controlsText, control.Key, control.Description)
	}
	app.Control.SetText(controlsText)
}
