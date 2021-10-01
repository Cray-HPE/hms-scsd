// MIT License
// 
// (C) Copyright [2018-2021] Hewlett Packard Enterprise Development LP
// 
// Permission is hereby granted, free of charge, to any person obtaining a
// copy of this software and associated documentation files (the "Software"),
// to deal in the Software without restriction, including without limitation
// the rights to use, copy, modify, merge, publish, distribute, sublicense,
// and/or sell copies of the Software, and to permit persons to whom the
// Software is furnished to do so, subject to the following conditions:
// 
// The above copyright notice and this permission notice shall be included
// in all copies or substantial portions of the Software.
// 
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
// THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
// OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
// ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
// OTHER DEALINGS IN THE SOFTWARE.

package hms_memcached

import (
	"fmt"
	"log"
	"context"
	"strconv"
	"strings"
	"time"
	"sync"
	"github.com/bradfitz/gomemcache/memcache"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/rest"
)


type serverInfo struct {
	podName string
	hostIP  string
	podIP   string
    port    int
}

type KV struct {
	Key   string `json:"Key"`
	Value string `json:"Value"`
}


/*type Memcached interface {
	Store(key string, value string) error
	Get(key string) (string, bool, error)
	GetMulti(key []string) ([]KV, error)
	Delete(key string) error
	TAS(key string) error
	Lock()
	Unlock()
	Close() error
}*/

var memcacheHandle *memcache.Client
var serverLock sync.Mutex
var serverList []serverInfo
var podBaseG string
var nameSpaceG string
var podPortG int
var scanMCRunning = false
var dieScan = false

func serverScan() error {
	fname := "serverScan()"
	config,err := rest.InClusterConfig()
	if (err != nil) {
		return fmt.Errorf("%s: clusterconfig failed: %v",fname,err)
	}

/*	kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config",)

    // Initialize kubernetes-client
    config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
    if err != nil {
        fmt.Printf("Error building kubeconfig: %v\n", err)
        os.Exit(1)
    }
*/

/*	kubeClient,cerr := kubernetes.NewForConfig(config)
	if (cerr != nil) {
		fmt.Errorf("%s: NewForConfig failed: %v",fname,cerr)
	}

	ctx,cancel := context.Background()
	pods,perr := clientset.CoreV1.Pods("services").List(ctx,metav1.ListOptions{})
	if (perr != nil) {
		log.Printf("%s: error getting pods: %v",fname,perr)
		return
	}
*/

	ctx := context.Background()
	kubeClient,cerr := kubernetes.NewForConfig(config)
	if (cerr != nil) {
		return fmt.Errorf("%s: NewForConfig failed: %v",fname,cerr)
	}
	podList, _ := kubeClient.CoreV1().Pods("services").List(ctx,metav1.ListOptions{})

	log.Printf("%s: Fetched info for %d pods.",fname,len(podList.Items))


	for _,pod := range(podList.Items) {
		log.Printf("%s:    podname:   '%s'",fname,pod.Name)
		log.Printf("%s:    namespace: '%s'",fname,pod.Namespace)
		log.Printf("%s:    phase:     '%s'",fname,pod.Status.Phase)
		log.Printf("%s:    podPort:   '%d'",fname,podPortG)
		log.Printf("%s:    hostIP:    '%s'",fname,pod.Status.HostIP)
		log.Printf("%s:    podIP:     '%s'",fname,pod.Status.PodIP)
		log.Printf("%s:    podIPs:")
		for _,ip := range(pod.Status.PodIPs) {
			log.Printf("%s:               '%s'",fname,ip.IP)
		}
		if (strings.Contains(pod.Name,podBaseG) && (pod.Namespace == nameSpaceG) &&
			(pod.Status.Phase == "Running")) {
			serverList = append(serverList,serverInfo{podName: pod.Name,
								hostIP: pod.Status.HostIP,
								podIP: pod.Status.PodIP,
								port: podPortG})
		}
	}

/* example output:
     podname:   'cray-hbtd-dff5969b5-24gh6'
     namespace: 'service'
     phase:     'running'
     hostIP:    '10.252.1.8'
     podIP:     '10.36.0.40'
     podIPs:
                '10.36.0.40'
     podname:   'cray-hbtd-dff5969b5-gpg69'
     namespace: 'service'
     phase:     'running'
     hostIP:    '10.252.1.7'
     podIP:     '10.47.0.50'
     podIPs:
                '10.47.0.50'
     podname:   'cray-hbtd-dff5969b5-p49dl'
     namespace: 'service'
     phase:     'running'
     hostIP:    '10.252.1.9'
     podIP:     '10.44.0.56'
     podIPs:
                '10.44.0.56'

//ZZZZ cray-hbtd-memcached
Thus, 'podname' is what to look for, and 'podIP' is the IP to use.
Q: what about a secondary container like a memcached server?
A: Gotta find it with this interface somehow.  And it's IP will be the same
   as the service container but will use a different port.

NOTE: Once the IP of other SCSD pods is known, that pod IP + memcached
      container port should enable communication with the memcached servers.
      Memcached container port is specified in the values.yaml file (32111?)
*/

	return nil
}

func updateMCServers() *memcache.Client {
	//Create server list string slice for memcache pkg.

	mcList := []string{}
	for _,sv := range(serverList) {
		svEntry := sv.podIP + ":" + strconv.Itoa(sv.port)
		mcList = append(mcList,svEntry)
	}

	client := memcache.New(mcList...)
	return client
}

func scanMC() {
	dly := 0

	if (scanMCRunning) {
		return
	}
	scanMCRunning = true
log.Printf("scanMC(): running")

	for {
		if (dieScan) {
			scanMCRunning = false
log.Printf("scanMC(): dying")
			return
		}

		time.Sleep(time.Duration(dly) * time.Second)
		dly = 10


		err := serverScan()
		if (err != nil) {
			log.Printf("%v",err)
			continue
		}

		if (len(serverList) == 0) {
			log.Printf("scanMC(): No memcached servers found.")
			continue
		}

		//At this point all the memcached servers are known.  Attach to them.

		serverLock.Lock()
		memcacheHandle = updateMCServers()
		serverLock.Unlock()
	}
}

func Open(podBase string, podPort int, nameSpace string) {
	serverLock.Lock()
	podBaseG = podBase
	podPortG = podPort
	nameSpaceG = nameSpace
	serverLock.Unlock()

	//Fire up the server scanner.  This will initialize the server connections.

	go scanMC()
}

func Store(key string, value string) error {
	data := memcache.Item{Key: key, Value: []byte(value)}
	serverLock.Lock()
	err := memcacheHandle.Set(&data)
	serverLock.Unlock()
	if (err != nil) {
		log.Printf("Store(): %v",err)
		return err
	}
	return nil
}

func StoreEphemeral(key string, value string, ttlSecs int32) error {
	data := memcache.Item{Key: key, Value: []byte(value), Expiration: ttlSecs}
	serverLock.Lock()
	err := memcacheHandle.Set(&data)
	serverLock.Unlock()
	if (err != nil) {
		log.Printf("StoreEphemeral(): %v",err)
		return err
	}
	return nil
}

func Touch(key string, ttlSecs int32) error {
	serverLock.Lock()
	err := memcacheHandle.Touch(key,ttlSecs)
	serverLock.Unlock()
	if (err != nil) {
		log.Printf("Touch(): %v",err)
		return err
	}
	return nil
}

func Exists(key string) bool {
	_,err := Get(key)
	if (err != nil) {
		return false
	}
	return true
}

func Get(key string) (string, error) {
	serverLock.Lock()
	item,err := memcacheHandle.Get(key)
	serverLock.Unlock()
	if (err != nil) {
		return "",err
	}

	return string(item.Value),nil
}

func GetMulti(keys []string) ([]KV, error) {
	var kvList []KV

	serverLock.Lock()
	kmap,err := memcacheHandle.GetMulti(keys)
	serverLock.Unlock()
	if (err != nil) {
		return kvList,err
	}

	for k,v := range(kmap) {
		kvList = append(kvList,KV{Key:k, Value: string(v.Value)})
	}

	return kvList,nil
}

func Delete(key string) error {
	serverLock.Lock()
	err := memcacheHandle.Delete(key)
	serverLock.Unlock()
	if (err != nil) {
		return err
	}
	return nil
}

func Close() {
	dieScan = true
	for {
		time.Sleep(1 * time.Millisecond)
		if (scanMCRunning == false) {
			break
		}
	}
	serverLock.Lock()
	memcacheHandle = nil
	serverLock.Unlock()
}


