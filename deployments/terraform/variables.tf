variable "region" {
    default = "us-east-1"
}

variable "profile" {
    default = "yaala"
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
    default = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQCyh1+7MKCRlgiGNuHDWN2teoqFr5KTGVk45xEywk4JiBTZjdkb27A6he4/tGfv7vAc4wvt3VhJDwzVOEwjGw+Qp3UJ7ZhYblS50Y2r+meaivZUVUDX3ghKTVHsCZaGotL3jGZDLwb87SzOCk7ZfaEinzX24VCzWriUgHCaSnLhAosh/wnj7ruWRgQfrQuZg4pW2PmalcpwjEwmnthUCH7n/j0+DJSyZXjVG0a8AdHFVh3UcdbP+c+BcMWcPgAGtpWeCn5ukEA0hQkXDEPGyP4RWTsAnPjR2bHHKBWcm85cvyxb6JHzHXJHsHnA9NI2cQ1E0FTWclKjqm71f1EJrXX1B53B7uofzZeGv38mj89xC6jXA09W90FDc0NuPfDk1HtpxunzV6FTAG4IGjXBX5pWOHZMIXM70MQE4bDi44zTrPt3RkZbfUhtu81rZkweTe45hXPsVHjoOnnbPYiNWNI/nlpg/3/xLvKTQLJBTS5KqVi7fqkld6fVflENZZu5bilwAPSeuRzZMsDKFoF04t/6VTus9B5F891IYQuY0jpoHJC+M36ZAR1DDuRXaaud7wg7yjVpZ+Vcq0msgvG1e37CrcFGIQZpMyS3dw9ZeZu/8vPPQX356vB64Lk9FmiKG/OIRpGGYj2+mTzSZ2Fn6Q4iH8fJjW0j098kj79wEYPmDQ== jayamal.devas@Jayamal-De-Vas-D1QPR94XFP"
}