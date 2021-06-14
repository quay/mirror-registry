FROM docker.io/centos:7

ENV ANSIBLE_VERSION 2.9.17

RUN yum check-update; \
    yum install -y gcc libffi-devel python3 epel-release; \
    yum install -y python3-pip; \
    yum install -y git; \
    yum clean all

RUN pip3 install --upgrade pip; \
    pip3 install "ansible==${ANSIBLE_VERSION}";

RUN git clone -b quay_installer_role https://github.com/quay/quay-ansible.git

ENV ANSIBLE_HOST_KEY_CHECKING False