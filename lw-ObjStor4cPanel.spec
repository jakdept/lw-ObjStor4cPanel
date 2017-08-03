%global lw_path /usr/local/lp/bin
%global bin_name lw-ObjStor4cPanel
%global git_repo github.com/jakdept/%{bin_name}
%global goversion 1.8.3
%global builddir ${RPM_BUILD_DIR}

Summary: Liquid Web CSF ranges
Name: %{bin_name}
Version: 1.0.0
Release: 0
License: MIT
Group: Applications/System
BuildRoot: %{_topdir}/%{name}-%{version}-%{release}-build
BuildArch: x86_64
Requires: bash
BuildRequires: curl git

%description
A transporter to connect cPanel's backup system to Liquid Web's Object Storage.

%prep
# install go
mkdir -p %{builddir}/go/{src,bin}
mkdir -p %{builddir}/usr/local

if ! go version ; then
  /usr/bin/curl -s -S -L \
    https://storage.googleapis.com/golang/go%{goversion}.linux-amd64.tar.gz|tar \
    xz -C %{builddir}/usr/local
fi

export PATH=%{builddir}/usr/local/go/bin:$PATH
export GOROOT=%{builddir}/usr/local/go
export GOPATH=%{builddir}/go

go get %{git_repo}
go get -t -v %{git_repo}/...

%build
export PATH=%{builddir}/usr/local/go/bin:$PATH
export GOROOT=%{builddir}/usr/local/go
export GOPATH=%{builddir}/go
export GOOS=linux
export GOARCH=amd64

go install %{git_repo}

%install
export PATH=%{builddir}/usr/local/go/bin:$PATH
export GOROOT=%{builddir}/usr/local/go
export GOPATH=%{builddir}/go

mkdir -p %{buildroot}/%{lw_path}
install -m 0755 ${GOPATH}/bin/%{bin_name} %{buildroot}%{lw_path}

%post
[[ $1 == 1 ]] && whmapi1 backup_destination_add \
  name=LW\ Object\ Storage \
  disabled=0 \
  type=Custom \
  upload_system_backup=on \
  script=%{lw_path}/%{bin_name} \
  host=bucketname \
  path=backups/ \
  timeout=300 \
  username=username \
  password=changeme > /dev/null

# cannot cleanly do preun action - cPanel assigns a random id and does not make
# it easy to find a specific backup destination

%clean
rm -rf ${RPM_BUILD_ROOT}

%files
%defattr(-,root,root)
%{lw_path}/%{bin_name}

%changelog
* Thu Aug  3 2017 Jack Hayhurst <jhayhurst@liquidweb.com> - version 1.0.0
- finished RPM and bumped version number for final deployment

* Thu Aug  3 2017 Jack Hayhurst <jhayhurst@liquidweb.com> - version 0.5.0
- included some fixes for secteam

* Wed Jul 31 2017 Jack Hayhurst <jhayhurst@liquidweb.com> - version 0.4.0
- renamed the script and rpm

* Thu Apr 13 2017 Jack Hayhurst <jhayhurst@liquidweb.com> - version 0.3
- got rpm version fully working - yay!

* Thu Apr 13 2017 Jack Hayhurst <jhayhurst@liquidweb.com> - version 0.2
- Wrote initial RPM