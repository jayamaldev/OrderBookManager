variable "region" {
    default = "us-east-1"
}

variable "availability_zone" {
    default = "us-east-1f"
}   

variable "instance_type" {
    default = "t2.medium"
}

variable "ami" {
    default = "ami-068c0051b15cdb816"
}

variable "key_name" {
    default = "ec2-key"
}   

variable "public_key" { //todo get from file
    default = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC/eElS32Pra85mffKm8XC7cy09edpOE2ijS6Ik/0/gHhtZ9j0JAoaiNZ8e3px5aSlj2GuM7Rs0kRMxXoXPhRyBTD9+iir/bRIaKTUiCGdHSyIioyMymAzFnuaCvDkNsrt+Tvk7ZCFGIQrsIYpD/e72AFFZ4v9jYaz+hqF/KmY4YUYOBsKKNkJCcvLEnOihXTuxe9UebmEwUmVnmU4xC7xHmzxmLgcU5hcZtbuxQK5nn5lOnJl78AgPxLyWQiHhC2c4UUXduwYOPnF14f9MiXNuYWuQp4KlzX5MKniTe286V6uVtYo3st3Uf/Ul4wUc/kW9wfwYbITjCPvCfmLKsG3GnioQEzsJApdqXtASLEnSUQ8eKwa75BoNeoPQ4L4MqAhgXTp/XuAO+CT29XoceKblIzbDoZpJHVxh4Pge4UG6wroSL1KwGt+5jTKKaz+raB7Su6ClhDXil8wU9Ej4lB4PVLzrWGShSOz5uOEGdFbJrkVyQv8gFgs18LT0PwENbkE= test@test.com"
}