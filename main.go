package main

import (
"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
  "time"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/majdif47/sysmetricslib/cpu"
	"github.com/majdif47/sysmetricslib/memory"

)

type model struct {
	Tabs       []string
	TabContent []string
	activeTab  int
  table       table.Model 
  cpuInfo    *cpu.CPUINFO
  memInfo     *memory.MemInfo
}


var (
	progressBarStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00"))
)
var (
	inactiveTabBorder = tabBorderWithBottom("┴", "─", "┴")
	activeTabBorder   = tabBorderWithBottom("┘", " ", "└")
	docStyle          = lipgloss.NewStyle()
	highlightColor    = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}
	inactiveTabStyle  = lipgloss.NewStyle().Border(inactiveTabBorder, true).BorderForeground(highlightColor).Padding(0, 1)
	activeTabStyle    = inactiveTabStyle.Border(activeTabBorder, true)
	windowStyle       = lipgloss.NewStyle().BorderForeground(highlightColor).Padding(2, 0).Align(lipgloss.Left).Border(lipgloss.NormalBorder()).UnsetBorderTop()
)



func (m model) Init() tea.Cmd {
	// Create a background task that periodically fetches both CPU and memory info
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		if m.activeTab == 0 {
      cpuInfo, err := cpu.GetCpuInfo()
      if err != nil {
        return fmt.Sprintf("Error: %v",err)
      }
      return cpuInfo
    } else if m.activeTab == 1 {
      memInfo, err := memory.GetMemoryStats()
      if err != nil {
        return fmt.Sprintf("Error: %v", err)
      }
      return memInfo
    }
    return nil
	})
}


func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "ctrl+c", "q":
			return m, tea.Quit
		case"tab":
			m.activeTab = min(m.activeTab+1, len(m.Tabs)-1)
			return m, tea.Tick(time.Second, func(t time.Time) tea.Msg {
				if m.activeTab == 1 {
					memInfo, err := memory.GetMemoryStats()
					if err != nil {
						return fmt.Sprintf("Error: %v", err)
					}
					return memInfo
				}
				return nil
			})
		case "shift+tab":
			m.activeTab = max(m.activeTab-1, 0)
			return m, nil
    case "k", "up":
      if m.activeTab == 0 {
        m.table.MoveUp(1)
        return m, nil
      }
    case "j", "down":
      if m.activeTab == 0 {
        m.table.MoveDown(1)
       return m, nil
      }
		}
	case *cpu.CPUINFO:
    m.TabContent[0] = fmt.Sprintf(
			"CPU: %s\nCores: %d\nThreads: %d\nCPU Usage: %s\nFrequency: %sMhz\n",
			msg.InfoCPU.ModelName, msg.InfoCPU.CoresCount, msg.InfoCPU.ThreadsCount, msg.UsageCPU, msg.InfoCPU.CurrentFreq,
		)
		m.cpuInfo = msg
		m.updateCpuTable(msg)
		return m, tea.Tick(time.Second, func(t time.Time) tea.Msg {
			cpuInfo, err := cpu.GetCpuInfo()
			if err != nil {
				return fmt.Sprintf("Error: %v", err)
			}
			return cpuInfo
		})
	case *memory.MemInfo:
		m.memInfo = msg
		m.updateMemoryContent(msg)
		return m, tea.Tick(time.Second, func(t time.Time) tea.Msg {
			memInfo, err := memory.GetMemoryStats()
			if err != nil {
				return fmt.Sprintf("Error: %v", err)
			}
			return memInfo
		})
	}

	return m, nil
}
func (m *model) updateCpuTable(cpuInfo *cpu.CPUINFO) {
	rows := make([]table.Row, 0, len(cpuInfo.TimePerThread))
	sortedThreads := make([]string, 0, len(cpuInfo.TimePerThread))
	for key := range cpuInfo.TimePerThread {
		sortedThreads = append(sortedThreads, key)
	}
	sortedThreads = SortThreads(sortedThreads)
	for _, thread := range sortedThreads {
		data := cpuInfo.TimePerThread[thread]
		rows = append(rows, table.Row{
      thread,
			fmt.Sprintf("%.2f%%", cpuInfo.UsageThreads[thread]),
			fmt.Sprintf("User: %d, System: %d, Idle: %d", data.User, data.System, data.Idle),
		})
	}
	m.table.SetRows(rows)
}


func (m *model) updateMemoryContent(memInfo *memory.MemInfo) {
	m.TabContent[1] = fmt.Sprintf(
		"Total: %.3fGB\nUsed: %.3fGB\nAvailable: %.3fGB\nSwap Total: %.3fGB\nSwap Used: %.3fGB\n\n%s\n%s\n%s\n",
		(memInfo.TotalMemory),
		(memInfo.UsedMemory),
		(memInfo.FreeMemory),
		(memInfo.TotalSwap),
		(memInfo.UsedSwap),
    renderBar(memInfo.UsedMemory, memInfo.TotalMemory, "Memory Usage:"),
    renderBar(memInfo.FreeMemory, memInfo.TotalMemory, "Available Memory:"),
    renderBar(memInfo.UsedSwap, memInfo.TotalSwap, "Swap Usage:"),
	)
}


func (m model) View() string {
	doc := strings.Builder{}

	// Render the tabs
	var renderedTabs []string
	for i, t := range m.Tabs {
		var style lipgloss.Style
		isFirst, isLast, isActive := i == 0, i == len(m.Tabs)-1, i == m.activeTab
		if isActive {
			style = activeTabStyle
		} else {
			style = inactiveTabStyle
		}
		border, _, _, _, _ := style.GetBorder()
		if isFirst && isActive {
			border.BottomLeft = "│"
		} else if isFirst && !isActive {
			border.BottomLeft = "├"
		} else if isLast && isActive {
			border.BottomRight = "│"
		} else if isLast && !isActive {
			border.BottomRight = "┤"
		}
		style = style.Border(border)
		renderedTabs = append(renderedTabs, style.Render(t))
	}
	row := lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)
	doc.WriteString(row)
	doc.WriteString("\n")
	activeTabContentStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#3E7B27")).Align(lipgloss.Left).Bold(true)
	if m.activeTab == 0 {
		doc.WriteString(windowStyle.Width((lipgloss.Width(row) - windowStyle.GetHorizontalFrameSize())).
			Render(
				activeTabContentStyle.Render(m.TabContent[m.activeTab]) + "\n\n" + m.table.View(),
			))
		doc.WriteString("\n")
	} else {
		doc.WriteString(windowStyle.Width((lipgloss.Width(row) - windowStyle.GetHorizontalFrameSize())).
			Render(activeTabContentStyle.Render(m.TabContent[m.activeTab])))
	}
	return docStyle.Render(doc.String())
}



func main() {
	columns := []table.Column{
		{Title: "Thread ID", Width: 20},
		{Title: "Usage (%)", Width: 20},
		{Title: "Time (ms)", Width: 50},
	}

  styles := table.DefaultStyles()
  styles.Header = lipgloss.NewStyle().Align(lipgloss.Center)
  styles.Cell = lipgloss.NewStyle().Align(lipgloss.Center)
  styles.Selected = lipgloss.NewStyle().Background(lipgloss.Color("#123524")).Bold(true).Align(lipgloss.Center).Foreground(lipgloss.Color("#EFE3C2"))
  t := table.New(
    table.WithColumns(columns),
    table.WithFocused(true), // Ensure focus is disabled
    table.WithStyles(styles),// Apply the updated styles
  )
	tabs := []string{"CPU", "Memory", "Running Tasks", "Disks", "Networks", "GPU", "General Info", "Power"}
	tabContent := []string{
		"Loading CPU info...",
		"Memory metrics...",
		"Task metrics...",
		"Disk metrics...",
		"Network metrics...",
		"GPU metrics...",
		"General system info...",
		"Power usage...",
	}

	m := model{
		Tabs:       tabs,
		TabContent: tabContent,
		table:      t,
	}

	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}


func renderBar(value, total float64, label string) string {
	percentage := float64(value) / float64(total)
  x := len("Available Memory:") - len(label)
  for x>0 {
    label += " "
    x--
  }
	bar := progressBarStyle.Render(strings.Repeat("█", int(percentage*20)) + strings.Repeat("░", 20-int(percentage*20)))
	return fmt.Sprintf("%s\t[%s] %.2f%%\n", label, bar, percentage*100)
}
func tabBorderWithBottom(left, middle, right string) lipgloss.Border {
	border := lipgloss.RoundedBorder()
	border.BottomLeft = left
	border.Bottom = middle
	border.BottomRight = right
	return border
}

func extractNumericSuffix(key string) int {
	// Find the numeric part of the key
	for i := len(key) - 1; i >= 0; i-- {
		if key[i] < '0' || key[i] > '9' {
			// Convert the numeric suffix to an integer
			num, _ := strconv.Atoi(key[i+1:])
			return num
		}
	}
	return 0 // Default value if no numeric suffix found
}

func SortThreads(keys []string) []string {
	sort.Slice(keys, func(i, j int) bool {
		// Extract numeric part from the keys
		num1 := extractNumericSuffix(keys[i])
		num2 := extractNumericSuffix(keys[j])

		return num1 < num2
	})
	return keys
}
