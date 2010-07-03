package core


type User struct {
	id   string
	nick string
}


func (u *User) Nick() string {
	c := make(chan string)
	corechan <- func() {
		c <- u.nick
	}
	return <-c
}

func (u *User) ID() string {
	c := make(chan string)
	corechan <- func() {
		c <- u.id
	}
	return <-c
}

func (u *User) Remove() {
	c := make(chan bool)
	corechan <- func() {
		users[u.id] = nil, false
		usersByNick[u.nick] = nil, false
		c <- true
	}
	<-c
}
