package main

import "fmt"
import "io/ioutil"
import "os"
import "strings"
import "sync"
import "runtime"
import "sort"
import "flag"

var wg sync.WaitGroup

func walk(path string, collect  chan<- string) {
    defer wg.Done()
    files, err := ioutil.ReadDir(path)
    if err != nil {
        return
    }

    if len(files) == 0 {
        collect <- path
        return
    }

    for _, file := range files {
        wg.Add(1)
        go func(file os.FileInfo, collect chan<- string) {
            defer wg.Done()
            if file.IsDir() {
                wg.Add(1)
                go walk(path + "/" + file.Name(), collect)
                return
            }
            collect <- path + "/" + file.Name()
        }(file, collect)
    }
}

func recursion_content(path string) []string {
    collect := make(chan string, 10000)
    var content []string
    go func(collect chan<- string) {
        wg.Add(1)
        go walk(path, collect)
        wg.Wait()
        close(collect)
    }(collect)
    for item := range collect {
        content = append(content, item)
    }
    return content
}

func non_recursion_content(path string) []string {
    files, err := ioutil.ReadDir(path)
    if err != nil {
        return []string{"error"}
    }

    var content []string
    for _, file := range files {
        content = append(content, file.Name())
    }
    return content
}

func main() {
    var folder_path, search_term string
    var ignore, recursion bool
    flag.StringVar(&folder_path, "p", ".", "path directory long to find")
    flag.StringVar(&search_term, "s", "*", "term long to search")
    flag.BoolVar(&ignore, "i", false, "ignore upper/lower")
    flag.BoolVar(&recursion, "r", false, "recursion find")
    flag.Parse()

    runtime.GOMAXPROCS(100)

    var content []string
    if recursion {
        content = recursion_content(folder_path)
    } else {
        content = non_recursion_content(folder_path)
    }
    sort.Strings(content)

    if search_term == "*" {
        for _, item := range content {
            fmt.Println(item)
        }
        return
    }

    compared_content := make([]string, len(content))
    copy(compared_content, content)
    var compared_st = search_term
    if ignore {
        compared_st = strings.ToLower(search_term)
        for i := range content {
            wg.Add(1)
            go func(idx int) {
                compared_content[idx] = strings.ToLower(content[idx])
                wg.Done()
            }(i)
        }
        wg.Wait()
    }

    for idx, item := range compared_content {
        if strings.Contains(item, compared_st) {
            fmt.Println(content[idx])
        }
    }
}
