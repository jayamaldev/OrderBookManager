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
        from_port = 80
        to_port = 80
        protocol = "tcp"
        cidr_blocks = ["0.0.0.0/0"]
    }

    ingress {
        from_port = 8080
        to_port = 8080
        protocol = "tcp"
        cidr_blocks = ["0.0.0.0/0"]
    }

    ingress {
        from_port = 22
        to_port = 22
        protocol = "tcp"
        cidr_blocks = ["10.0.1.0/24"]
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
    security_groups = [aws_security_group.jayamal_bastion_sg.id]
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
    security_groups = [aws_security_group.jayamal_app_sg.id]
    instance_type = var.instance_type
    key_name = aws_key_pair.ec2_key_pair.key_name
    user_data = <<-EOF
    #!/bin/bash
    yum update -y
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
    subnets = [aws_subnet.jayamal_ob_public_subnet.id]

    tags = {
      Name = "jayamal_lb"
      Author = "Jayamal"
    }
}

resource "aws_lb_target_group" "jayamal_app_tg_8080" {
    name = "jayamal-app-tg-8080"
    port = 8080
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
    port = 8080
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

resource "aws_key_pair" "ec2_key_pair" {
    key_name = var.key_name
    public_key = var.public_key
}

output "order_book_public_url" {
    value = aws_lb.jayamal_lb.dns_name
}

output "bastion_host_ip" {
    value = aws_instance.bastion_host.public_ip
}