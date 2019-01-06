#!/bin/bash


install_go() {
    if ! which go
    then
        pushd /tmp
        go_package='https://dl.google.com/go/go1.11.2.linux-amd64.tar.gz'
        wget "$go_package"
        tar -C /usr/local -xzf $(echo "$go_package" | cut -f5 -d/)
        ls -la /usr/local/go
        popd
    else
        echo "golang found"
        go --version
    fi

    mkdir -p ~/go/{pkg,bin,src}

    echo "go get -u golang.org/x/tools/cmd/..."
    go get -u golang.org/x/tools/cmd/...
    echo "go get -u github.com/rogpeppe/godef/..."
    go get -u github.com/rogpeppe/godef/...
    echo "go get -u github.com/nsf/gocode"
    go get -u github.com/nsf/gocode
    echo "go get -u golang.org/x/tools/cmd/goimports"
    go get -u golang.org/x/tools/cmd/goimports
    echo "go get -u golang.org/x/tools/cmd/guru"
    go get -u golang.org/x/tools/cmd/guru
    echo "go get -u github.com/dougm/goflymake"
    go get -u github.com/dougm/goflymake
    echo "All done, enjoy!"

}

install_shellcheck() {
    if ! which shellcheck &> /dev/null
    then
        sudo apt install shellcheck || echo "Could not download shellcheck" ; exit 1
    else
        echo "shellcheck found."
    fi
}

###############
# Install git #
###############
install_git() {
    if ! which git &> /dev/null
    then
        apt install git -y
    else
        echo "git found."
    fi
}
##############
# Find emacs #
##############
install_emacs() {
    if ! emacs --version | head -1 | awk '{print $3}' | grep "^25."
    then
        echo "Installing emacs"
    else
        echo "$(emacs --version | head -1) found."
    fi
}
###################
# configure_emacs #
###################
configure_emacs(){
    if ! [ -d ~/.emacs.d/.git ]
    then
        echo "rm ~/.emacs.d" && rm -rf ~/.emacs.d
        git clone https://github.com/r4wm/.emacs.d ~/.emacs.d
    fi
}

update_bashrc(){
    cp ".bashrc" ~/
    source ~/.bashrc
}


install_git
install_emacs
configure_emacs
install_shellcheck
update_bashrc
install_go

