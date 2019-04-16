package user_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/pedrocelso/go-task/lib/http/authcontext"
	"github.com/pedrocelso/go-task/lib/services/user"
	"github.com/stretchr/testify/assert"

	"fmt"

	"cloud.google.com/go/datastore"
	"google.golang.org/appengine/aetest"
)

const email = `pedro@pedrocelso.com.br`

var mainCtx authcontext.Context
var c context.Context
var client *datastore.Client

func TestMain(m *testing.M) {
	ctx, done, _ := aetest.NewContext()
	c := context.Background()
	client, _ := datastore.NewClient(c, "go-rest-client")
	mainCtx.DataStoreClient = client
	mainCtx.AppEngineCtx = ctx
	_ = createUsers(mainCtx)
	os.Exit(m.Run())
	done()
}

func TestCreateUser(t *testing.T) {
	output, err := user.Create(&mainCtx, &user.User{
		Name:  `Pedro Costa`,
		Email: `pedro@pedrocelso.com.br`,
	})

	assert.Nil(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, "Pedro Costa", output.Name)
	assert.Equal(t, "pedro@pedrocelso.com.br", output.Email)

	output, err = user.Create(&mainCtx, nil)
	assert.NotNil(t, err)
	assert.Equal(t, "error: invalid User data", err.Error())
	assert.Nil(t, output)

	output, err = user.Create(&mainCtx, &user.User{
		Name:  `Pedro Costa`,
		Email: ``,
	})

	assert.NotNil(t, err)
	assert.Equal(t, "error: invalid User data", err.Error())
	assert.Nil(t, output)
}

func TestGetByEmail(t *testing.T) {
	err := createUsers(mainCtx)
	if err != nil {
		t.Fatal(err)
	}

	output, err := user.GetByEmail(&mainCtx, `pedro@pedrocelso.com.br1`)
	assert.Nil(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, "Pedro 1", output.Name)
	assert.Equal(t, "pedro@pedrocelso.com.br1", output.Email)

	output, err = user.GetByEmail(&mainCtx, `bad_email@gmail.com`)
	assert.NotNil(t, err)
	assert.Equal(t, "user 'bad_email@gmail.com' not found", err.Error())
	assert.Nil(t, output)

	output, err = user.GetByEmail(&mainCtx, ``)
	assert.NotNil(t, err)
	assert.Equal(t, `error: invalid User data`, err.Error())
	assert.Nil(t, output)
}

// This test run on a different context ot ensure that only
// the created users will be saved on the datastore
func TestGetUsers(t *testing.T) {
	var authCtx authcontext.Context
	ctx, done, err := aetest.NewContext()
	authCtx.AppEngineCtx = ctx
	defer done()
	err = createUsers(authCtx)
	if err != nil {
		t.Fatal(err)
	}
	// This sleep is needed because it take some milliseconds for the objects
	// created on `createUsers` to be indexed and returned on query
	time.Sleep(time.Millisecond * 5e2)
	output, err := user.GetUsers(&authCtx)
	assert.Nil(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, 5, len(output))
}

func TestUpdateUser(t *testing.T) {
	err := createUsers(mainCtx)

	output, err := user.Update(&mainCtx, &user.User{
		Name:  `Migeh`,
		Email: `pedro@pedrocelso.com.br0`,
	})
	assert.Nil(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, "Migeh", output.Name)
	assert.Equal(t, "pedro@pedrocelso.com.br0", output.Email)

	usr, err := user.GetByEmail(&mainCtx, `pedro@pedrocelso.com.br0`)
	assert.Nil(t, err)
	assert.NotNil(t, output)
	assert.Equal(t, "Migeh", usr.Name)
	assert.Equal(t, "pedro@pedrocelso.com.br0", usr.Email)

	output, err = user.Update(&mainCtx, nil)
	assert.NotNil(t, err)
	assert.Equal(t, "error: invalid User data", err.Error())
	assert.Nil(t, output)
}

func TestDeleteUser(t *testing.T) {
	err := createUsers(mainCtx)

	usr, err := user.GetByEmail(&mainCtx, `pedro@pedrocelso.com.br0`)
	assert.Nil(t, err)
	assert.NotNil(t, usr)
	assert.Equal(t, "Pedro 0", usr.Name)
	assert.Equal(t, "pedro@pedrocelso.com.br0", usr.Email)

	err = user.Delete(&mainCtx, `pedro@pedrocelso.com.br0`)
	assert.Nil(t, err)

	usr, err = user.GetByEmail(&mainCtx, `pedro@pedrocelso.com.br0`)
	assert.NotNil(t, err)
	assert.Equal(t, "user 'pedro@pedrocelso.com.br0' not found", err.Error())
	assert.Nil(t, usr)
}

func createUsers(ctx authcontext.Context) error {
	for i := 0; i < 5; i++ {
		email := fmt.Sprintf(`%v%v`, email, i)
		name := fmt.Sprintf(`Pedro %v`, i)
		key := datastore.NameKey(`User`, email, nil)
		if _, err := client.Put(ctx.AppEngineCtx, key, &user.User{
			Name:  name,
			Email: email,
		}); err != nil {
			return err
		}
	}
	return nil
}
