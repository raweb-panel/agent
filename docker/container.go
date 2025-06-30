package docker

import (
    "context"
    "encoding/json"
    "net/http"
    "runtime"
    "strings"
    "sync"
    "time"

    "github.com/docker/docker/api/types"
    "github.com/docker/docker/api/types/container"
    "github.com/docker/docker/api/types/mount"
    "github.com/docker/docker/api/types/network"
    "github.com/docker/docker/client"
)

var (
    statsMutex sync.RWMutex
    statsCache = make(map[string]*containerStats)
)

type containerStats struct {
    CPUUsage    uint64
    SystemUsage uint64
    MemoryUsage uint64
    MemoryLimit uint64
    NetworkRx uint64
    NetworkTx uint64
    Timestamp time.Time
}

type DeleteRequest struct {
    ID string `json:"id"`
}

type ActionRequest struct {
    ID string `json:"id"`
}

type CreateContainerRequest struct {
    Image   string                 `json:"image"`
    Name    string                 `json:"name"`
    Volumes []map[string]string    `json:"volumes"`
    Labels  map[string]string      `json:"labels"`
    IPv4    string                 `json:"ipv4"`
    IPv6    string                 `json:"ipv6"`
}

func ListContainers() ([]types.Container, error) {
    cli, err := client.NewClientWithOpts(client.WithHost(dockerHost), client.WithAPIVersionNegotiation())
    if err != nil {
        return nil, err
    }
    defer cli.Close()

    containers, err := cli.ContainerList(context.Background(), container.ListOptions{All: true})
    if err != nil {
        return nil, err
    }
    return containers, nil
}

func ListContainersHandler(w http.ResponseWriter, r *http.Request) {
    containers, err := ListContainers()
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
        return
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(containers)
}

func DeleteContainerHandler(w http.ResponseWriter, r *http.Request) {
    var req DeleteRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ID == "" {
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(map[string]string{"error": "Missing or invalid container id"})
        return
    }

    cli, err := client.NewClientWithOpts(client.WithHost(dockerHost), client.WithAPIVersionNegotiation())
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
        return
    }
    defer cli.Close()

    opts := container.RemoveOptions{
        RemoveVolumes: false,
        RemoveLinks:   false,
        Force:         true,
    }

    err = cli.ContainerRemove(context.Background(), req.ID, opts)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{"message": "Container deleted"})
}

func StopContainerHandler(w http.ResponseWriter, r *http.Request) {
    var req ActionRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ID == "" {
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(map[string]string{"error": "Missing or invalid container id"})
        return
    }
    cli, err := client.NewClientWithOpts(client.WithHost(dockerHost), client.WithAPIVersionNegotiation())
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
        return
    }
    defer cli.Close()
    if err := cli.ContainerStop(context.Background(), req.ID, container.StopOptions{}); err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
        return
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{"message": "Container stopped"})
}

func StartContainerHandler(w http.ResponseWriter, r *http.Request) {
    var req ActionRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ID == "" {
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(map[string]string{"error": "Missing or invalid container id"})
        return
    }
    cli, err := client.NewClientWithOpts(client.WithHost(dockerHost), client.WithAPIVersionNegotiation())
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
        return
    }
    defer cli.Close()
    if err := cli.ContainerStart(context.Background(), req.ID, container.StartOptions{}); err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
        return
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{"message": "Container started"})
}

func KillContainerHandler(w http.ResponseWriter, r *http.Request) {
    var req ActionRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ID == "" {
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(map[string]string{"error": "Missing or invalid container id"})
        return
    }
    cli, err := client.NewClientWithOpts(client.WithHost(dockerHost), client.WithAPIVersionNegotiation())
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
        return
    }
    defer cli.Close()
    if err := cli.ContainerKill(context.Background(), req.ID, "SIGKILL"); err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
        return
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{"message": "Container killed"})
}

func CreateContainerHandler(w http.ResponseWriter, r *http.Request) {
    var req CreateContainerRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request"})
        return
    }

    cli, err := client.NewClientWithOpts(client.WithHost(dockerHost), client.WithAPIVersionNegotiation())
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
        return
    }
    defer cli.Close()

    var mounts []mount.Mount
    for _, v := range req.Volumes {
        mode := mount.TypeBind
        rw := true
        if v["mode"] == "ro" {
            rw = false
        }
        mounts = append(mounts, mount.Mount{
            Type:   mode,
            Source: v["host"],
            Target: v["container"],
            ReadOnly: !rw,
        })
    }

    networkingConfig := &network.NetworkingConfig{}
    endpointsConfig := make(map[string]*network.EndpointSettings)
    ipamConfig := &network.EndpointIPAMConfig{}
    if req.IPv4 != "" {
        ipamConfig.IPv4Address = req.IPv4
    }
    if req.IPv6 != "" {
        ipamConfig.IPv6Address = req.IPv6
    }
    if req.IPv4 != "" || req.IPv6 != "" {
        endpointsConfig["bridge"] = &network.EndpointSettings{
            IPAMConfig: ipamConfig,
        }
        networkingConfig.EndpointsConfig = endpointsConfig
    }

    resp, err := cli.ContainerCreate(
        context.Background(),
        &container.Config{
            Image:  req.Image,
            Labels: req.Labels,
        },
        &container.HostConfig{
            Mounts: mounts,
        },
        networkingConfig,
        nil,
        req.Name,
    )
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
        return
    }

    if err := cli.ContainerStart(context.Background(), resp.ID, container.StartOptions{}); err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "status": "started",
        "id":     resp.ID,
    })
}

func GetContainerByIDHandler(w http.ResponseWriter, r *http.Request) {
    id := r.URL.Query().Get("id")
    if id == "" {
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(map[string]string{"error": "Missing container id"})
        return
    }

    cli, err := client.NewClientWithOpts(client.WithHost(dockerHost), client.WithAPIVersionNegotiation())
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
        return
    }
    defer cli.Close()

    containerJSON, err := cli.ContainerInspect(context.Background(), id)
    if err != nil {
        w.WriteHeader(http.StatusNotFound)
        json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(containerJSON)
}

func GetContainerByNameHandler(w http.ResponseWriter, r *http.Request) {
    var name string

    if r.Method == http.MethodGet {
        name = r.URL.Query().Get("name")
    } else if r.Method == http.MethodPost {
        var body struct {
            Name string `json:"name"`
        }
        if err := json.NewDecoder(r.Body).Decode(&body); err == nil {
            name = body.Name
        }
    }

    if name == "" {
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(map[string]string{"error": "Missing container name"})
        return
    }

    cli, err := client.NewClientWithOpts(client.WithHost(dockerHost), client.WithAPIVersionNegotiation())
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
        return
    }
    defer cli.Close()

    containers, err := cli.ContainerList(context.Background(), container.ListOptions{All: true})
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
        return
    }

    for _, c := range containers {
        for _, n := range c.Names {
            if strings.TrimPrefix(n, "/") == name {
                containerJSON, err := cli.ContainerInspect(context.Background(), c.ID)
                if err != nil {
                    w.WriteHeader(http.StatusNotFound)
                    json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
                    return
                }
                w.Header().Set("Content-Type", "application/json")
                json.NewEncoder(w).Encode(containerJSON)
                return
            }
        }
    }

    w.WriteHeader(http.StatusNotFound)
    json.NewEncoder(w).Encode(map[string]string{"error": "Container not found"})
}

func GetContainerStatsByNameHandler(w http.ResponseWriter, r *http.Request) {
    var name string
    if r.Method == http.MethodGet {
        name = r.URL.Query().Get("name")
    } else if r.Method == http.MethodPost {
        var body struct {
            Name string `json:"name"`
        }
        if err := json.NewDecoder(r.Body).Decode(&body); err == nil {
            name = body.Name
        }
    }
    if name == "" {
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(map[string]string{"error": "Missing container name"})
        return
    }

    cli, err := client.NewClientWithOpts(client.WithHost(dockerHost), client.WithAPIVersionNegotiation())
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
        return
    }
    defer cli.Close()

    containerID := ""
    containers, err := cli.ContainerList(context.Background(), container.ListOptions{All: true})
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
        return
    }

    for _, c := range containers {
        for _, n := range c.Names {
            if strings.TrimPrefix(n, "/") == name {
                containerID = c.ID
                break
            }
        }
    }

    if containerID == "" {
        w.WriteHeader(http.StatusNotFound)
        json.NewEncoder(w).Encode(map[string]string{"error": "Container not found"})
        return
    }

    containerInfo, err := cli.ContainerInspect(context.Background(), containerID)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
        return
    }

    statsResp, err := cli.ContainerStatsOneShot(context.Background(), containerID)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
        return
    }
    defer statsResp.Body.Close()

    var stats container.Stats
    if err := json.NewDecoder(statsResp.Body).Decode(&stats); err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
        return
    }

    hostCPUs := len(stats.CPUStats.CPUUsage.PercpuUsage)
    if hostCPUs == 0 {
        hostCPUs = runtime.NumCPU()
    }

    cpuPercent := calculateCPUPercentage(containerID, stats)
    memUsage := float64(stats.MemoryStats.Usage) / (1024 * 1024)
    memLimit := float64(stats.MemoryStats.Limit) / (1024 * 1024)
    networkRx, networkTx := calculateNetworkUsage(containerID, stats)
    cpuLimitPercent := float64(hostCPUs * 100)
    if containerInfo.HostConfig.NanoCPUs > 0 {
        cpuLimitPercent = float64(containerInfo.HostConfig.NanoCPUs) / 10000000
    } else if containerInfo.HostConfig.CPUQuota > 0 && containerInfo.HostConfig.CPUPeriod > 0 {
        cpuLimitPercent = float64(containerInfo.HostConfig.CPUQuota) / float64(containerInfo.HostConfig.CPUPeriod) * 100
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "cpu_percent": cpuPercent,
        "cpu_limit_percent": cpuLimitPercent,
        "mem_usage_mb": memUsage,
        "mem_limit_mb": memLimit,
        "network_rx_bytes": networkRx,
        "network_tx_bytes": networkTx,
        "host_cpus": hostCPUs,
    })
}

func calculateCPUPercentage(containerID string, stats container.Stats) float64 {
    statsMutex.Lock()
    defer statsMutex.Unlock()

    now := time.Now()
    currentStats := &containerStats{
        CPUUsage:    stats.CPUStats.CPUUsage.TotalUsage,
        SystemUsage: stats.CPUStats.SystemUsage,
        MemoryUsage: stats.MemoryStats.Usage,
        MemoryLimit: stats.MemoryStats.Limit,
        Timestamp:   now,
    }

    if stats.Networks != nil {
        for _, network := range stats.Networks {
            currentStats.NetworkRx += network.RxBytes
            currentStats.NetworkTx += network.TxBytes
        }
    }

    previousStats, exists := statsCache[containerID]
    statsCache[containerID] = currentStats

    if !exists || now.Sub(previousStats.Timestamp) > 30*time.Second {
        numCPUs := len(stats.CPUStats.CPUUsage.PercpuUsage)
        if numCPUs == 0 {
            numCPUs = runtime.NumCPU()
        }
        if stats.CPUStats.SystemUsage > 0 {
            return (float64(stats.CPUStats.CPUUsage.TotalUsage) / float64(stats.CPUStats.SystemUsage)) *
                   float64(numCPUs) * 100.0
        }
        return 0.0
    }
    cpuDelta := float64(currentStats.CPUUsage - previousStats.CPUUsage)
    systemDelta := float64(currentStats.SystemUsage - previousStats.SystemUsage)
    if systemDelta > 0 && cpuDelta > 0 {
        numCPUs := len(stats.CPUStats.CPUUsage.PercpuUsage)
        if numCPUs == 0 {
            numCPUs = runtime.NumCPU()
        }
        return (cpuDelta / systemDelta) * float64(numCPUs) * 100.0
    }
    return 0.0
}
func calculateNetworkUsage(containerID string, stats container.Stats) (uint64, uint64) {
    rx := uint64(0)
    tx := uint64(0)

    if stats.Networks != nil {
        for _, network := range stats.Networks {
            rx += network.RxBytes
            tx += network.TxBytes
        }
    }

    return rx, tx
}
