package cache

import (
	"fmt"
	"log"
	"sync"

	aiven "github.com/aiven/aiven-go-client"
)

var (
	acls         = make(map[string]map[string]aiven.KafkaACL)
	aclCacheLock sync.Mutex
)

//ACLCache type
type ACLCache struct {
}

//Read populates the cache if it doesn't exist, and reads the required acl. An aiven.Error with status
//404 is returned upon cache miss
func (a ACLCache) Read(project, service, aclID string, client *aiven.Client) (acl aiven.KafkaACL, err error) {
	aclCacheLock.Lock()
	defer aclCacheLock.Unlock()
	if _, ok := acls[project+service]; !ok {
		if err = a.populateACLCache(project, service, client); err != nil {
			return
		}
	}
	if cachedService, ok := acls[project+service]; ok {
		if acl, ok = cachedService[aclID]; !ok {
			// cache miss, try to get the ACL from the Aiven API instead
			log.Printf("Cache miss on ACL: %s, going live to Aiven API", aclID)
			var liveACL *aiven.KafkaACL
			if liveACL, err = client.KafkaACLs.Get(project, service, aclID); err == nil {
				acl = *liveACL
			}
		}
	} else {
		//TODO: returning a 404 here provides no extra value, as the ACL is then treated as if it
		// doesn't exist (which may not be the case here).
		err = aiven.Error{
			Status:  404,
			Message: fmt.Sprintf("Cache miss on project/service: %s/%s", project, service),
		}
	}

	return
}

//write writes the specified ACL to the cache
func (a ACLCache) write(project, service string, acl *aiven.KafkaACL) (err error) {
	var cachedService map[string]aiven.KafkaACL
	var ok bool
	if cachedService, ok = acls[project+service]; !ok {
		cachedService = make(map[string]aiven.KafkaACL)
	}

	cachedService[acl.ID] = *acl
	acls[project+service] = cachedService
	return
}

//Refresh refreshes the ACL cache
func (a ACLCache) Refresh(project, service string, client *aiven.Client) error {
	aclCacheLock.Lock()
	defer aclCacheLock.Unlock()
	return a.populateACLCache(project, service, client)
}

//populateACLCache makes a call to Aiven to list kafka ACLs, and upserts into the cache
func (a ACLCache) populateACLCache(project, service string, client *aiven.Client) (err error) {
	var acls []*aiven.KafkaACL
	if acls, err = client.KafkaACLs.List(project, service); err == nil {
		for _, acl := range acls {
			a.write(project, service, acl)
		}
	}
	return
}
