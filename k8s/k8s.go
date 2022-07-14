package k8s

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/trivago/tgo/tcontainer"

	"github.com/joho/godotenv"
	jira "gopkg.in/andygrunwald/go-jira.v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var exclusions = []string{
	"test1",
}

// Check if given value is in slice
func contains(slice []string, key string) bool {
	for _, str := range slice {
		if key == str {
			return true
		}
	}
	return false
}

func main() {
	ns := flag.String("namespace", "", "namespace")
	parent := flag.String("parentid", "", "parentid")
	spoints := flag.String("spoints", "", "spoints")
	labels := []string{"EKS", "Neodymium"}

	flag.Parse()
	godotenv.Load(".env")
	// Create a BasicAuth Transport object
	tp := jira.BasicAuthTransport{
		Username: os.Getenv("JIRA_USER"),
		Password: os.Getenv("JIRA_TOKEN"),
	}
	// Create a new Jira Client
	client, err := jira.NewClient(tp.Client(), os.Getenv("JIRA_URL"))
	if err != nil {
		log.Fatal(err)
	}

	// Bootstrap k8s configuration from local 	Kubernetes config file
	kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
	log.Println("Using kubeconfig file: ", kubeconfig)

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Fatal(err)
	}

	// Create an rest client not targeting specific API version
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	list, err := clientset.AppsV1().Deployments(*ns).List(context.Background(), metav1.ListOptions{})

	if err != nil {
		log.Fatalln("failed to get deployments:", err)
	}

	unknowns := tcontainer.NewMarshalMap()
	unknowns["customfield_10002"] = spoints

	for i, deployment := range list.Items {
		fmt.Printf("[%d] %s\n", i, deployment.GetName())

		if contains(exclusions, deployment.GetName()) {
			fmt.Printf("Skipping: %v", deployment.GetName())
			continue
		}

		j := jira.Issue{
			Fields: &jira.IssueFields{
				Reporter: &jira.User{
					AccountID: "5a69d1e8187efa232b09cf1f",
				},
				Summary:     fmt.Sprintf("Implement Right Sizing Methodology for %v", deployment.GetName()),
				Description: fmt.Sprintf("Implement Right Sizing Methodology for %v", deployment.GetName()),
				Project: jira.Project{
					Key: "NEO",
				},
				Type: jira.IssueType{
					Name: "Task",
				},
				Parent: &jira.Parent{
					ID: *parent,
				},
				Labels:   labels,
				Unknowns: unknowns,
			},
		}
		issue, resp, err := client.Issue.Create(&j)
		if err != nil {
			body, _ := ioutil.ReadAll(resp.Body)
			fmt.Println(string(body))
			panic(err)
		}
		fmt.Printf("%s: %+v\n", issue.Key, issue.Self)

	}
}
