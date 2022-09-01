package dependency_test

import (
	"github.com/golang/mock/gomock"
	"github.com/shshimamo/Hands-On-Software-Engineering-with-Golang/ch4test/dependency"
	mock_dependency "github.com/shshimamo/Hands-On-Software-Engineering-with-Golang/ch4test/dependency/mock"
	"reflect"
	"testing"
)

func TestDependencyCollector(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	api := mock_dependency.NewMockAPI(ctrl)

	gomock.InOrder(
		api.EXPECT().
			ListDependencies("proj0").
			Return([]string{"proj1", "res1"}, nil),
		api.EXPECT().
			DependencyType("proj1").
			Return(dependency.DepTypeProject, nil),
		api.EXPECT().
			DependencyType("res1").
			Return(dependency.DepTypeResource, nil),
		api.EXPECT().
			ListDependencies("proj1").
			Return([]string{"res1", "res2"}, nil),
		api.EXPECT().
			DependencyType("res2").
			Return(dependency.DepTypeResource, nil),
	)

	collector := dependency.NewCollector(api)
	depList, err := collector.AllDependencies("proj0")
	if err != nil {
		t.Fatal(err)
	}

	if exp := []string{"proj1", "res1", "res2"}; !reflect.DeepEqual(depList, exp) {
		t.Fatalf("expected dependency list to be:\n%v\ngot:\n%v", exp, depList)
	}
}