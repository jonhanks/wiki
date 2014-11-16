package main

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestAnonymousUser(t *testing.T) {
	Convey("The UserInfo struct defaults to an anonymous user when not initialized", t, func() {
		var u *UserInfo

		So(u.IsAnonymous(), ShouldBeTrue)

		Convey("An uninitilized or Anonymous user should have no roles", func() {
			So(len(u.Roles()), ShouldBeZeroValue)

			Convey("Adding a role to an uninitialized or Anonymous user will do nothing", func() {
				u.AddRole("Admin")
				So(len(u.Roles()), ShouldBeZeroValue)
			})
		})

		Convey("Test with an allocated by empty UserInfo", func() {
			u = &UserInfo{}
			Convey("An uninitilized or Anonymous user should have no roles", func() {
				So(len(u.Roles()), ShouldBeZeroValue)

				Convey("Adding a role to an uninitialized or Anonymous user will do nothing", func() {
					u.AddRole("Admin")
					So(len(u.Roles()), ShouldBeZeroValue)
				})
			})
		})
	})
}

func TestUser(t *testing.T) {
	Convey("An initialized user has a username", t, func() {
		u := &UserInfo{username: "AdminUser"}

		So(u.IsAnonymous(), ShouldBeFalse)
		So(u.Username(), ShouldEqual, "AdminUser")

		Convey("Before adding roles there should be no roles", func() {
			So(len(u.Roles()), ShouldBeZeroValue)

			Convey("Adding a role should bring the count to 1", func() {
				u.AddRole("Admin")
				So(len(u.Roles()), ShouldEqual, 1)
				So(u.Roles(), ShouldContain, "Admin")

				Convey("Adding duplicate roles does nothing", func() {
					u.AddRole("Admin")
					So(len(u.Roles()), ShouldEqual, 1)
					So(u.Roles(), ShouldContain, "Admin")
				})
			})
		})
	})
}
