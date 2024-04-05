package mock

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/testing"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	stvziov1 "stvz.io/coral/pkg/apis/stvz.io/v1"
)

type Client struct {
	log     logr.Logger
	tracker testing.ObjectTracker
	scheme  *runtime.Scheme
	client.Client
}

// NewClient returns a mock (fake) client for testing. The fixtures are
// not automatically loaded into the cache.  Individual fixtures can be loaded
// using the WithFixtureOrDie method and all fixtures in a directory can be loaded
// using LoadAllOrDie.
func NewClient() *Client {
	s := scheme.Scheme
	_ = stvziov1.AddToScheme(s)

	tracker := testing.NewObjectTracker(s, scheme.Codecs.UniversalDecoder())
	client := fake.NewClientBuilder().
		WithObjectTracker(tracker).
		WithScheme(s).
		WithStatusSubresource(&stvziov1.Image{}).
		Build()

	return &Client{
		log:     logr.Discard(),
		scheme:  s,
		tracker: tracker,
		Client:  client,
	}
}

// WithLogger sets the logger for the client.
func (m *Client) WithLogger(log logr.Logger) *Client {
	m.log = log
	return m
}

// WithFixtureOrDie loads a single fixture into the cache.  The fixture must be in a
// recognizable format for the universal deserializer.
func (m *Client) WithFixtureOrDie(filename ...string) *Client {
	decoder := scheme.Codecs.UniversalDeserializer()
	for _, f := range filename {
		data, err := os.ReadFile(f)
		if err != nil {
			panic(err)
		}

		sections := strings.Split(string(data), "---")

		for _, section := range sections {
			data = []byte(section)
			obj, _, err := decoder.Decode(data, nil, nil)
			if err != nil {
				panic(err)
			}

			// Fake some of the creation metadata.  There's probably a few other
			// things that could be useful.
			obj.(client.Object).SetCreationTimestamp(metav1.Time{
				Time: metav1.Now().Time,
			})

			err = m.tracker.Add(obj)
			if err != nil {
				panic(err)
			}
		}
	}
	return m
}

// LoadAllOrDie loads all fixtures in the directory into the cache.
func (m *Client) LoadAllOrDie(dir string) *Client {
	files, err := filepath.Glob(dir + "/*.yaml")
	if err != nil {
		panic(err)
	}

	m.WithFixtureOrDie(files...)

	return m
}
