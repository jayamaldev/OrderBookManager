resource "aws_iam_role" "ob_ec2_role_jayamal" {
    name = "ob_ec2_role_jayamal"

    assume_role_policy = jsonencode({
        Version = "2012-10-17"
        Statement = [
        {
            Action = "sts:AssumeRole"
            Effect = "Allow"
            Sid    = ""
            Principal = {
            Service = "ec2.amazonaws.com"
            }
        },
        ]
    })

    tags = {
        Name = "ob_ec2_role_jayamal"
        Owner = "jayamal"
    }
}

resource "aws_iam_role_policy_attachment" "ecr_read_only_jayamal" {
    role = aws_iam_role.ob_ec2_role_jayamal.name
    policy_arn = "arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly"
}

resource "aws_iam_instance_profile" "ob_ec2_profile_jayamal" {
    name = "ob_ec2_profile_jayamal"
    role = aws_iam_role.ob_ec2_role_jayamal.name
}

resource "aws_vpc" "jayamal_ob_vpc" {
    cidr_block = "10.0.0.0/16"

    tags = {
      Name = "jayamal_ob_vpc"
      Owner = "jayamal"
    }
}

resource "aws_internet_gateway" "jayamal_int_gw" {
    vpc_id = aws_vpc.jayamal_ob_vpc.id

    tags = {
      Name = "Jayamal Internet Gateway"
      Author = "Jayamal"
    }
}

resource "aws_subnet" "jayamal_ob_public_subnet" {
    vpc_id = aws_vpc.jayamal_ob_vpc.id
    cidr_block = "10.0.1.0/24"
    availability_zone = var.availability_zone
    
    tags = {
        Name = "Jayamal Public Subnet"
        Owner = "jayamal"
    }
}

resource "aws_subnet" "jayamal_ob_private_subnet" {
    vpc_id = aws_vpc.jayamal_ob_vpc.id
    cidr_block = "10.0.2.0/24"
    availability_zone = var.availability_zone

    tags = {
        Name = "Jayamal Private Subnet"
        Owner = "jayamal"
    }
}

resource "aws_eip" "jayamal_elastic_ip" {
    domain = "vpc"

    tags = {
      Name = "Jayamal Elastic IP"
      Owner = "Jayamal"
    }
}

resource "aws_nat_gateway" "jayamal_nat_gw" {
    allocation_id = aws_eip.jayamal_elastic_ip.id
    subnet_id = aws_subnet.jayamal_ob_public_subnet.id

    tags = {
      Name = "Jayamal NAT GW"
      Author = "Jayamal"
    }
}

resource "aws_route_table" "jayamal_public_route_table" {
    vpc_id = aws_vpc.jayamal_ob_vpc.id

    route {
        cidr_block = "0.0.0.0/0"
        gateway_id = aws_internet_gateway.jayamal_int_gw.id
    }

    tags = {
      Name = "Jayamal Route Table"
      Author = "Jayamal"
    }
}

resource "aws_route_table_association" "jayamal_public_subnet_assoc" {
    subnet_id = aws_subnet.jayamal_ob_public_subnet.id
    route_table_id = aws_route_table.jayamal_public_route_table.id
}

resource "aws_route_table" "jayamal_private_route_table" {
    vpc_id = aws_vpc.jayamal_ob_vpc.id

    route {
        cidr_block = "0.0.0.0/0"
        nat_gateway_id = aws_nat_gateway.jayamal_nat_gw.id
    }

    tags = {
        Name = "Jayamal Private Route Table"
        Author = "Jayamal"
    }
}

resource "aws_route_table_association" "jayamal_private_subnet_assoc" {
    subnet_id = aws_subnet.jayamal_ob_private_subnet.id
    route_table_id = aws_route_table.jayamal_private_route_table.id
}

resource "aws_security_group" "jayamal_lb_sg" {
    name = "jayamal-lb-sg"
    vpc_id = aws_vpc.jayamal_ob_vpc.id

    ingress {
        from_port = 8080
        to_port = 8080
        protocol = "tcp"
        cidr_blocks = ["0.0.0.0/0"]
    }

    egress {
        from_port = 0
        to_port = 0
        protocol = "-1"
        cidr_blocks = ["0.0.0.0/0"]
    }

    tags = {
      Name = "jayamal_lb_sg"
      Author = "Jayamal"
    }
}

resource "aws_security_group" "jayamal_bastion_sg" {
    name = "jayamal-bastion-sg"
    vpc_id = aws_vpc.jayamal_ob_vpc.id

    ingress {
        from_port = 22
        to_port = 22
        protocol = "tcp"
        cidr_blocks = ["0.0.0.0/0"]
    }

    egress {
        from_port = 0
        to_port = 0
        protocol = "-1"
        cidr_blocks = ["0.0.0.0/0"]
    }

    tags = {
      Name = "jayamal_bastion_sg"
      Author = "Jayamal"
    }
}

resource "aws_security_group" "jayamal_app_sg" {
    name = "jayamal_app_sg"
    vpc_id = aws_vpc.jayamal_ob_vpc.id

    ingress {
        from_port = 30088
        to_port = 30088
        protocol = "tcp"
        security_groups = [aws_security_group.jayamal_lb_sg.id]
        # cidr_blocks = ["0.0.0.0/0"]
    }

    ingress {
        from_port = 22
        to_port = 22
        protocol = "tcp"
        # cidr_blocks = ["10.0.1.0/24"]
        security_groups = [aws_security_group.jayamal_bastion_sg.id]
    }

    ingress { // to allow internal self routes for k8s
        from_port = 0
        to_port = 0
        protocol = "-1"
        self = true
    }

    egress {
        from_port = 0
        to_port = 0
        protocol = "-1"
        cidr_blocks = ["0.0.0.0/0"]
    }

    tags = {
      Name = "jayamal_app_sg"
      Author = "Jayamal"
    }
}

resource "aws_instance" "bastion_host" {
    ami = var.ami
    subnet_id = aws_subnet.jayamal_ob_public_subnet.id
    instance_type = var.instance_type
    vpc_security_group_ids = [aws_security_group.jayamal_bastion_sg.id]
    associate_public_ip_address = true
    key_name = aws_key_pair.ec2_key_pair.key_name

    tags = {
      Name = "bastion_host"
      Owner = "Jayamal"
    }
}

resource "aws_instance" "order_book_app" {
    ami = var.ami
    subnet_id = aws_subnet.jayamal_ob_private_subnet.id
    vpc_security_group_ids = [aws_security_group.jayamal_app_sg.id]
    instance_type = var.instance_type
    key_name = aws_key_pair.ec2_key_pair.key_name
    iam_instance_profile = aws_iam_instance_profile.ob_ec2_profile_jayamal.name
    user_data = <<-EOF
    #!/bin/bash
    yum update -y
    wget -O /etc/yum.repos.d/snapd.repo https://bboozzoo.github.io/snapd-amazon-linux/al2023/snapd.repo
    dnf install snapd -y
    systemctl enable --now snapd.socket
    ln -s /var/lib/snapd/snap /snap
    yum install httpd -y
    systemctl start httpd
    systemctl enable httpd
    echo "<html><h1>Hello World!</h1></html>" > /var/www/html/index.html
  EOF

    tags = {
      Name = "order-book-app"
      Owner = "Jayamal"
    }
}

resource "aws_lb" "jayamal_lb" {
    name = "jayamal-lb"
    internal = false
    load_balancer_type = "network"
    security_groups = [aws_security_group.jayamal_lb_sg.id]
    subnets = [aws_subnet.jayamal_ob_public_subnet.id]

    tags = {
      Name = "jayamal_lb"
      Author = "Jayamal"
    }
}

resource "aws_lb_target_group" "jayamal_app_tg_8080" {
    name = "jayamal-app-tg-8080"
    port = 30088
    protocol = "TCP"
    vpc_id = aws_vpc.jayamal_ob_vpc.id

    tags = {
      Name = "jayamal_app_tg-8080"
      Author = "Jayamal"
    }
}

resource "aws_lb_target_group" "jayamal_app_tg_80" {
    name = "jayamal-app-tg-80"
    port = 80
    protocol = "TCP"
    vpc_id = aws_vpc.jayamal_ob_vpc.id

    tags = {
      Name = "jayamal_app_tg_80"
      Author = "Jayamal"
    }
}

resource "aws_lb_target_group_attachment" "jayamal_lb_attach_8080" {
    target_group_arn = aws_lb_target_group.jayamal_app_tg_8080.arn
    target_id = aws_instance.order_book_app.id
    port = 30088
}

resource "aws_lb_target_group_attachment" "jayamal_lb_attach_80" {
    target_group_arn = aws_lb_target_group.jayamal_app_tg_80.arn
    target_id = aws_instance.order_book_app.id
    port = 80
}

resource "aws_lb_listener" "jayamal_lb_listner" {
    load_balancer_arn = aws_lb.jayamal_lb.arn
    port = 80
    protocol = "TCP"
    default_action {
      type = "forward"
      target_group_arn = aws_lb_target_group.jayamal_app_tg_80.arn
    }

    tags = {
      Name = "jayamal_lb_listner"
      Author = "Jayamal"
    }
}

resource "aws_lb_listener" "jayamal_lb_listner_8080" {
    load_balancer_arn = aws_lb.jayamal_lb.arn
    port = 8080
    protocol = "TCP"
    default_action {
      type = "forward"
      target_group_arn = aws_lb_target_group.jayamal_app_tg_8080.arn
    }

    tags = {
      Name = "jayamal_lb_listner_8080"
      Author = "Jayamal"
    }
}

# resource "aws_key_pair" "ec2_key_pair" {
#     key_name = var.key_name 
#     public_key = file("~/.ssh/id_rsa.pub") 
# }

resource "aws_key_pair" "ec2_key_pair" {
    key_name = var.key_name
    public_key = var.public_key
}

resource "aws_ecr_repository" "jayamal_ecr_repo" {
  name                 = "jayamal_ecr_repo"
  image_tag_mutability = "IMMUTABLE"

  image_scanning_configuration {
    scan_on_push = true
  }

  tags = {
      Name = "jayamal_ecr_repo"
      Author = "Jayamal"
    }
}

output "order_book_public_url" {
    value = aws_lb.jayamal_lb.dns_name
}

output "bastion_host_ip" {
    value = aws_instance.bastion_host.public_ip
}

output "order_book_private_ip" {
    value = aws_instance.order_book_app.private_ip
}