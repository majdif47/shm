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
	"github.com/majdif47/sysmetricslib/disks"
	"github.com/majdif47/sysmetricslib/memory"
	"github.com/majdif47/sysmetricslib/networks"
	
)

type model struct {
	Tabs       []string
	TabContent []string
	activeTab  int
  CPUtable    table.Model 
  NetTabel    table.Model
  cpuInfo    *cpu.CPUINFO
  memInfo     *memory.MemInfo
  disksInfo   *disks.DiskStats
  networkStats []networks.NetInfo
  width        int
  height       int
}


var (
	progressBarStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#00ADB5")).Bold(true)
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
    } else if m.activeTab== 2 {
      disksInfo, err := disks.GetDiskUsage()
      if err != nil {
        return fmt.Sprintf("Error: %v", err)
      }
      return disksInfo
    } else if m.activeTab == 3 {
        networkStats, err := networks.GetNetworkInterfaces()
        if err != nil {
          return fmt.Sprintf("Error: %v", err)
        }
        return networkStats
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
				}else if m.activeTab == 2 {
	    		disksInfo, err := disks.GetDiskUsage()
			    if err != nil {
				  return fmt.Sprintf("Error: %v", err)
			  }
			  return disksInfo
        }else if m.activeTab == 3 {
          netStats, err := networks.GetNetworkInterfaces()
          if err != nil {
            return fmt.Sprintf("Error: %v", err)
          }
          return netStats
        }
				return nil
			})
		case "shift+tab":
			m.activeTab = max(m.activeTab-1, 0)
			return m, nil
    case "k", "up":
      if m.activeTab == 0 {
        m.CPUtable.MoveUp(1)
        return m, nil
      }else if m.activeTab == 3{
        m.NetTabel.MoveUp(1)
        return m,nil
      }
    case "j", "down":
      if m.activeTab == 0 {
        m.CPUtable.MoveDown(1)
       return m, nil
      } else if m.activeTab == 3 {
        m.NetTabel.MoveDown(1)
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
  case *disks.DiskStats:
    m.disksInfo = msg
    m.updateDiskContent(msg)
    return m,tea.Tick(time.Second, func(t time.Time) tea.Msg {
			disksInfo, err := disks.GetDiskUsage()
			if err != nil {
				return fmt.Sprintf("Error: %v", err)
			}
			return disksInfo
		})
  case []networks.NetInfo:
    m.TabContent[3] = fmt.Sprintf("Network Stats: \n")
    m.networkStats = msg
    m.updateNetworkTable(msg)
    return m, tea.Tick(time.Second, func(t time.Time) tea.Msg {
      netInfo, err := networks.GetNetworkInterfaces()
      if err != nil {
        return fmt.Sprintf("Error: %v", err)
      }
      return netInfo
    })
  case tea.WindowSizeMsg:
    m.height = msg.Height
    m.width = msg.Width
    return m, nil
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
	m.CPUtable.SetRows(rows)
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

func (m *model) updateDiskContent(diskInfo *disks.DiskStats) {
	m.TabContent[2] = fmt.Sprintf(
		"Total: %.3fGB\nUsed: %.3fGB\nAvailable: %.3fGB\n\n%s\n%s\n",
		float64(diskInfo.Total)/1e9,
		float64(diskInfo.Used)/1e9,
		float64(diskInfo.Free)/1e9,
    renderBar(float64(diskInfo.Used), float64(diskInfo.Total), "Disk Usage:"),
    renderBar(float64(diskInfo.Free), float64(diskInfo.Total), "Free Space:"),
	)
}


func (m *model) updateNetworkTable(netInfo []networks.NetInfo) {
  rows := make([]table.Row, 0, len(netInfo))
  for i := range netInfo {
    data := netInfo[i]
    rows = append(rows, table.Row{
      data.Name,
      data.State,
      fmt.Sprintf("%d",data.Speed),
      fmt.Sprintf("%d",data.RxBytes),
      fmt.Sprintf("%d",data.TxBytes),
      fmt.Sprintf("%d",data.RxErrors),
      fmt.Sprintf("%d",data.TxErrors),
    })
  }
  m.NetTabel.SetRows(rows)
}

func (m model) View() string {
	doc := strings.Builder{}
  tabWidth := m.width/len(m.Tabs)
  for i := 0 ; i < len(m.Tabs); i++ {
    tabWidth -= len(m.Tabs[i])
  }
  whiteSpace := ""
  for i := 0 ; i <= tabWidth ; i++ {
    whiteSpace += " "
  }
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
    t += whiteSpace
		renderedTabs = append(renderedTabs, style.Render(t))
	}
	row := lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)
	doc.WriteString(row)
	doc.WriteString("\n")
	activeTabContentStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#DEFCF9")).Align(lipgloss.Left).Bold(true)
	if m.activeTab == 0 {
    m.CPUtable.Focus()
		doc.WriteString(windowStyle.Width((lipgloss.Width(row) - windowStyle.GetHorizontalFrameSize())).
			Render(
				activeTabContentStyle.Render(m.TabContent[m.activeTab]) + "\n\n" + m.CPUtable.View(),
			))
		doc.WriteString("\n")
	} else if m .activeTab == 3 {
    m.NetTabel.Focus()
		doc.WriteString(windowStyle.Width((lipgloss.Width(row) - windowStyle.GetHorizontalFrameSize())).
			Render(
				activeTabContentStyle.Render(m.TabContent[m.activeTab]) + "\n\n" + m.NetTabel.View(),
			))
  }else {
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
  
  columns1 := []table.Column{
    {Title: "Interface\t", Width: 10},
    {Title: "State",Width: 10},
    {Title: "Speed", Width: 10},
    {Title: "RxBytes", Width: 15},
    {Title: "TxBytes", Width: 15},
    {Title: "RxErrors", Width: 10},
    {Title: "TxErrors", Width: 10},
  }
  styles := table.DefaultStyles()
  styles.Header = lipgloss.NewStyle().Align(lipgloss.Center)
  styles.Cell = lipgloss.NewStyle().Align(lipgloss.Center)
  styles.Selected = lipgloss.NewStyle().Bold(true).Align(lipgloss.Center).Foreground(lipgloss.Color("#00ADB5"))
  t := table.New(
    table.WithColumns(columns),
    table.WithStyles(styles),// Apply the updated styles
  )
  t1 := table.New(
    table.WithColumns(columns1),
    table.WithStyles(styles),
    )
	tabs := []string{"CPU", "Memory", "Disks", "Networks","General Info"}
	tabContent := []string{
		"Loading CPU info...",
		"Memory metrics...",
		"Disk metrics...",
		"Network metrics...",
		"General system info...",
	}
	m := model{
		Tabs:       tabs,
		TabContent: tabContent,
		CPUtable:   t,
    NetTabel:   t1, 
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
  str := (fmt.Sprintf("%s\t%s %.2f%%\n", label, bar, percentage*100))
	return progressBarStyle.Render(str)
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
