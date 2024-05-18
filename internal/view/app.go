package view

import (
	"fmt"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func NewApp(chainURL string, chainID uint64) (*App, error) {
	app := &App{
		Application: tview.NewApplication(),
		chainURL:    chainURL,
		chainID:     chainID,
		fullTopView: tview.NewFlex().SetDirection(tview.FlexRow),
		ContextView: NewContextView(chainURL, chainID),
		MainView:    NewMainView(),
	}

	app.ContextView.SetBorder(true)
	app.ContextView.SetChangedFunc(func() {
		app.Draw()
	})

	app.ControlView = tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWordWrap(true).
		SetWrap(true).
		SetChangedFunc(func() {
			app.Draw()
		})
	app.ControlView.SetBorder(true)

	app.contextAndControlView = tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(app.ContextView, 0, 1, false).
		AddItem(app.ControlView, 0, 1, false)
	app.fullTopView.AddItem(app.contextAndControlView, 0, 1, false)

	app.fullView = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(app.fullTopView, 0, 1, false).
		AddItem(app.MainView, 0, 4, true)

	app.defaultControls = ControlMapping{
		NormalControls: NormalKeyControls{
			'/': Control{
				Key:         "/",
				Description: "Search",
				Fn:          Search,
			},
		},
		SpecialControls: SpecialKeyControls{
			tcell.KeyCtrlC: Control{
				Key:         "Ctrl+C",
				Description: "Quit",
				Fn: func(app *App) error {
					app.Stop()
					return nil
				},
			},
		},
	}
	app.SetRoot(app.fullView, true)

	client, err := ethclient.Dial(chainURL)
	if err != nil {
		return nil, fmt.Errorf("Error connecting to chain at %s: %s", chainURL, err.Error())
	}

	sum, err := NewChainSummary(app, client)
	if err != nil {
		return nil, fmt.Errorf("Error getting chain summary: %s", err.Error())
	}

	app.Update(sum)
	return app, nil
}

type App struct {
	*tview.Application

	chainURL string
	chainID  uint64

	fullView              *tview.Flex
	contextAndControlView *tview.Flex
	fullTopView           *tview.Flex

	ContextView     *ContextView
	ControlView     *tview.TextView
	defaultControls ControlMapping
	MainView        *MainView
	currentView     View
}

func (app *App) Update(view View) {
	if app.currentView != nil {
		app.currentView.End()
	}

	app.currentView = view
	app.MainView.Clear()
	app.MainView.AddItem(view.Show(), 0, 1, false)
	app.UpdateControls(view.Controls())
}

type ContextView struct {
	*tview.TextView

	permanentText string
}

func NewContextView(chainURL string, chainID uint64) *ContextView {
	permanentText := fmt.Sprintf("URL: %s\nID: %d", chainURL, chainID)
	t := tview.NewTextView().
		SetText(permanentText).
		SetDynamicColors(true).
		SetRegions(true).
		SetWordWrap(true).
		SetWrap(true)
	return &ContextView{
		TextView:      t,
		permanentText: permanentText,
	}
}

func (c *ContextView) AddError(errorText string) {
	c.Clear()
	c.SetText(fmt.Sprintf("%s\n[red]Error: %s[-]", c.permanentText, errorText))
}

func (c *ContextView) AddContext(contextText string) {
	c.Clear()
	c.SetText(fmt.Sprintf("%s\n[yellow]%s[-]", c.permanentText, contextText))
}

type MainView struct {
	*tview.Flex
}

func NewMainView() *MainView {
	return &MainView{
		Flex: tview.NewFlex(),
	}
}

func (app *App) UpdateControls(controls ControlMapping) {
	if controls.NormalControls == nil {
		controls.NormalControls = NormalKeyControls{}
	}
	if controls.SpecialControls == nil {
		controls.SpecialControls = SpecialKeyControls{}
	}
	for key, defaultControl := range app.defaultControls.NormalControls {
		controls.NormalControls[key] = defaultControl
	}
	for key, defaultControl := range app.defaultControls.SpecialControls {
		controls.SpecialControls[key] = defaultControl
	}
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyRune:
			if control, exists := controls.NormalControls[event.Rune()]; exists {
				if control.Fn == nil {
					app.ContextView.AddError(fmt.Sprintf("No function for control %c", event.Rune()))
					return nil
				}
				if err := control.Fn(app); err != nil {
					app.ContextView.AddError(err.Error())
					return nil
				}
			}
		default:
			if control, exists := controls.SpecialControls[event.Key()]; exists {
				if control.Fn == nil {
					return nil
				}
				if err := control.Fn(app); err != nil {
					app.ContextView.AddError(err.Error())
					return nil
				}
			}
		}
		return event
	})

	controlsText := ""
	for _, control := range controls.NormalControls {
		controlsText = fmt.Sprintf("%s[blue]%s[-] : %s\n", controlsText, control.Key, control.Description)
	}
	for _, control := range controls.SpecialControls {
		controlsText = fmt.Sprintf("%s[blue]%s[-] : %s\n", controlsText, control.Key, control.Description)
	}
	app.ControlView.SetText(controlsText)
}
