package main

import (
	"fmt"
	"strings"
)

// Edge 表示一条有向边。
type Edge struct {
	To     int // 目标节点编号
	Weight int // 边的权重，例如距离或通行成本
}

// 使用 int 能表示的最大值，代表节点暂时不可达。
const INF = int(^uint(0) >> 1)

// Dijkstra 计算起点 start 到所有节点的最短距离。
//
// graph：邻接表，graph[u] 保存从节点 u 出发的所有边。
// dist：dist[v] 表示起点到节点 v 的最短距离。
// prev：prev[v] 表示最短路径上，节点 v 的前一个节点。
func Dijkstra(graph [][]Edge, start int) ([]int, []int, error) {
	n := len(graph)

	if start < 0 || start >= n {
		return nil, nil, fmt.Errorf("起点编号越界：%d", start)
	}

	// 检查输入：Dijkstra 不支持负权边。
	for from, edges := range graph {
		for _, edge := range edges {
			if edge.To < 0 || edge.To >= n {
				return nil, nil, fmt.Errorf(
					"节点 %d 的目标节点越界：%d", from, edge.To,
				)
			}
			if edge.Weight < 0 {
				return nil, nil, fmt.Errorf(
					"边 %d -> %d 的权重不能为负数", from, edge.To,
				)
			}
		}
	}

	dist := make([]int, n)
	prev := make([]int, n)
	visited := make([]bool, n)

	// 初始化：所有节点距离为无穷大，前驱节点为 -1。
	for i := 0; i < n; i++ {
		dist[i] = INF
		prev[i] = -1
	}

	// 起点到自己的距离为 0。
	dist[start] = 0

	for count := 0; count < n; count++ {
		// 第一步：在未处理的节点中，找到当前距离最小的节点 u。
		u := -1

		for v := 0; v < n; v++ {
			if !visited[v] && (u == -1 || dist[v] < dist[u]) {
				u = v
			}
		}

		// 没有可处理的节点，或者剩余节点都不可达，提前结束。
		if u == -1 || dist[u] == INF {
			break
		}

		// 因为边权非负，此时 u 的最短距离已确定。
		visited[u] = true

		// 第二步：尝试通过 u，缩短到其邻居的距离。
		// 这个更新过程叫作“松弛”。
		for _, edge := range graph[u] {
			v := edge.To

			if visited[v] {
				continue
			}

			// 防止整数相加溢出。
			// INF 被用作不可达标记，因此也不用于表示有限距离。
			if edge.Weight >= INF-dist[u] {
				continue
			}

			newDist := dist[u] + edge.Weight

			// 如果经过 u 更近，就更新距离和路径。
			if newDist < dist[v] {
				dist[v] = newDist
				prev[v] = u
			}
		}
	}

	return dist, prev, nil
}

// buildPath 根据前驱数组，还原起点到目标节点的路径。
// 这里的 prev 来自 Dijkstra；调用前已确认目标可达。
func buildPath(prev []int, target int) []int {
	var path []int

	// 从终点沿着前驱节点，倒着走回起点。
	for current := target; current != -1; current = prev[current] {
		path = append(path, current)
	}

	// 将倒序路径翻转为：起点 -> ... -> 终点。
	for left, right := 0, len(path)-1; left < right; left, right = left+1, right-1 {
		path[left], path[right] = path[right], path[left]
	}

	return path
}

func main() {
	// 使用编号表示节点：
	// A=0、B=1、C=2、D=3、E=4、F=5。
	names := []string{"A", "B", "C", "D", "E", "F"}

	// 有向图的邻接表。
	// 例如 A -> B 权重为 4，A -> C 权重为 1。
	graph := [][]Edge{
		{{To: 1, Weight: 4}, {To: 2, Weight: 1}}, // A
		{{To: 3, Weight: 1}},                     // B
		{{To: 1, Weight: 2}, {To: 3, Weight: 5}}, // C
		{{To: 4, Weight: 3}},                     // D
		{},                                       // E
		{},                                       // F：与其他节点不连通
	}

	start := 0 // 从 A 出发

	dist, prev, err := Dijkstra(graph, start)
	if err != nil {
		fmt.Println("计算失败：", err)
		return
	}

	for target := range graph {
		if dist[target] == INF {
			fmt.Printf("%s -> %s：不可达\n", names[start], names[target])
			continue
		}

		path := buildPath(prev, target)
		pathNames := make([]string, len(path))

		for i, node := range path {
			pathNames[i] = names[node]
		}

		fmt.Printf(
			"%s -> %s：最短距离 = %d，路径 = %s\n",
			names[start],
			names[target],
			dist[target],
			strings.Join(pathNames, " -> "),
		)
	}
}
