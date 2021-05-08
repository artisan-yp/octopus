package kubectl

import (
	"errors"
	"flag"
	"fmt"
	"time"

	"github.com/k8s-practice/octopus/xlog"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	r1 "k8s.io/apimachinery/pkg/runtime"
	r2 "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/workqueue"
)

var (
	logger     = xlog.Component("kubectl")
	kubeconfig = flag.String("kubeconfig", "", "kubernetes config file path.")
	master     = flag.String("master", "", "kubernetes cluster(https://hostname:port).")
)

const (
	RESOURCE_PODS      = "pods"
	RESOURCE_ENDPOINTS = "endpoints"
)

type SubscribeFunc func(name, obj interface{})

type Controller struct {
	indexer  cache.Indexer
	informer cache.Controller
	queue    workqueue.RateLimitingInterface

	resource string
	callback SubscribeFunc
}

func New(namespace string,
	resource string,
	selector fields.Selector,
	stopWatch chan struct{},
	callback SubscribeFunc,
) (*Controller, error) {
	cfg, err := clientcmd.BuildConfigFromFlags(*master, *kubeconfig)
	if err != nil {
		logger.Errorln(err)
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		logger.Errorln(err)
		return nil, err
	}

	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	objType, err := getObjectType(resource)
	if err != nil {
		return nil, err
	}
	listWatch := cache.NewListWatchFromClient(clientset.CoreV1().RESTClient(),
		resource, namespace, selector)
	indexer, informer := cache.NewIndexerInformer(listWatch,
		objType,
		0,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				key, err := cache.MetaNamespaceKeyFunc(obj)
				if err == nil {
					queue.Add(key)
					logger.Infof("Add %s, key: %s\n", resource, key)
				} else {
					logger.Errorf("Add %s, error: %v\n", resource, err)
				}
			},
			UpdateFunc: func(oldObj interface{}, newObj interface{}) {
				key, err := cache.MetaNamespaceKeyFunc(newObj)
				if err == nil {
					queue.Add(key)
					logger.Infof("Update %s, key: %s\n", resource, key)
				} else {
					logger.Errorf("Update %s, error: %v\n", resource, err)
				}
			},
			DeleteFunc: func(obj interface{}) {
				key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
				if err == nil {
					queue.Add(key)
					logger.Infof("Delete %s, key: %s\n", resource, key)
				} else {
					logger.Infof("Delete %s, error: %v\n", resource, err)
				}
			},
		},
		cache.Indexers{},
	)

	return &Controller{
		indexer:  indexer,
		informer: informer,
		queue:    queue,
		resource: resource,
		callback: callback,
	}, nil
}

func Subscribe(namespace string,
	resource string,
	selector fields.Selector,
	stopWatch chan struct{},
	callback SubscribeFunc,
) error {
	c, err := New(namespace, resource, selector, stopWatch, callback)
	if err != nil {
		logger.Errorln(err)
		return err
	}

	go c.Run(1, stopWatch)

	return nil
}

func (c *Controller) Run(threadiness int, stopWatch chan struct{}) {
	defer r2.HandleCrash()

	defer c.queue.ShutDown()
	logger.Infof("Starting Endpoints Controller...")

	go c.informer.Run(stopWatch)

	if !cache.WaitForCacheSync(stopWatch, c.informer.HasSynced) {
		r2.HandleError(fmt.Errorf("Failed sync caches."))
		return
	}

	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopWatch)
	}

	<-stopWatch
	logger.Infoln("Stoping Endpoints Controller...")
}

func (c *Controller) runWorker() {
	for c.processNextItem() {
	}
}

func (c *Controller) processNextItem() bool {
	key, quit := c.queue.Get()
	if quit {
		return false
	}

	defer c.queue.Done(key)

	err := c.sync(key.(string))
	c.HandleErr(err, key)

	return true
}

func (c *Controller) sync(key string) error {
	item, exists, err := c.indexer.GetByKey(key)
	if err != nil {
		logger.Errorf("key: %s, error: %v\n", key, err)
		return err
	} else if !exists {
		logger.Warningf("Source %s does not exist anymore\n", key)
	}

	switch c.resource {
	case RESOURCE_ENDPOINTS:
		endpoints, ok := item.(*v1.Endpoints)
		if !ok {
			logger.Warningf("Resource is %s, but item is not %s.",
				RESOURCE_ENDPOINTS, RESOURCE_ENDPOINTS)
		}
		return c.syncEndpoints(endpoints)
	case RESOURCE_PODS:
		pods, ok := item.(*v1.Pod)
		if !ok {
			logger.Warningf("Resource is %s, but item is not a %s.",
				RESOURCE_PODS, RESOURCE_PODS)
		}
		return c.syncPods(pods)
	default:
		return nil
	}
}

func (c *Controller) HandleErr(err error, key interface{}) {
	if err == nil {
		c.queue.Forget(key)
		return
	}

	if c.queue.NumRequeues(key) < 5 {
		logger.Errorf("Error sync resource %v: %v\n", key, err)
		c.queue.AddRateLimited(key)
		return
	}

	c.queue.Forget(key)

	r2.HandleError(err)
	logger.Errorf("Dropping resource %q out of the queue: %v", key, err)
}

func getObjectType(resource string) (r1.Object, error) {
	switch resource {
	case RESOURCE_ENDPOINTS:
		return &v1.Endpoints{}, nil
	case RESOURCE_PODS:
		return &v1.Pod{}, nil
	default:
		return nil, errors.New(fmt.Sprintf("Unknown resource: %s.", resource))
	}
}

func (c *Controller) syncEndpoints(endpoints *v1.Endpoints) error {
	if endpoints == nil {
		c.callback(nil, nil)
		return nil
	}

	logger.Infof("Sync endpoint %s", endpoints.GetName())
	ips := make([]string, 0)
	for _, subset := range endpoints.Subsets {
		for _, address := range subset.Addresses {
			ips = append(ips, address.IP)
			logger.Infof("%s %s\n", endpoints.GetName(), address.IP)
		}
	}

	c.callback(endpoints.GetName(), ips)
	return nil
}

func (c *Controller) syncPods(*v1.Pod) error {
	// TODO
	return nil
}
