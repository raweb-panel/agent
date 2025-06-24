package docker

import (
    "context"
    "encoding/json"
    "net/http"

    "github.com/docker/docker/api/types"
    "github.com/docker/docker/api/types/container"
    "github.com/docker/docker/client"
)

type DeleteRequest struct {
    ID string `json:"id"`
}

type ActionRequest struct {
    ID string `json:"id"`
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
