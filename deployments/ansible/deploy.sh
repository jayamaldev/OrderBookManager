#!/bin/bash

SSH_KEY_PATH="$HOME/.ssh/id_rsa"
ANSIBLE_INVENTORY_FILE="hosts.ini"
ANSIBLE_PLAYBOOK_FILE="setup.yaml"

echo "Starting Order Book App Ansible Playbook"

eval $(ssh-agent -s)
ssh-add $SSH_KEY_PATH

ssh-add -l

ansible-playbook -i $ANSIBLE_INVENTORY_FILE $ANSIBLE_PLAYBOOK_FILE