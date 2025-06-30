package docker

import (
    "context"
    "encoding/json"
    "net/http"

    "github.com/docker/docker/api/types/image"
    "github.com/docker/docker/client"
)

func ListImages() ([]image.Summary, error) {
    cli, err := client.NewClientWithOpts(client.WithHost(dockerHost), client.WithAPIVersionNegotiation())
    if err != nil {
        return nil, err
    }
    defer cli.Close()

    images, err := cli.ImageList(context.Background(), image.ListOptions{All: true})
    if err != nil {
        return nil, err
    }
    return images, nil
}

func ListImagesHandler(w http.ResponseWriter, r *http.Request) {
    images, err := ListImages()
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
        return
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(images)
}

func DeleteImageHandler(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Registry string `json:"registry"`
    }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Registry == "" {
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(map[string]string{"error": "Missing or invalid registry"})
        return
    }

    cli, err := client.NewClientWithOpts(client.WithHost(dockerHost), client.WithAPIVersionNegotiation())
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
        return
    }
    defer cli.Close()

    _, err = cli.ImageRemove(context.Background(), req.Registry, image.RemoveOptions{Force: true, PruneChildren: true})
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{"message": "Image deleted"})
}
