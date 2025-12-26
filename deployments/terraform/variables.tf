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
    default = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQC3v92JxIAlEmQqUk5anDcfhVRKLsJSgeA69OwLKEb3KGF1FG9/P7y5C15My/4/Yw5SST3JFqaMRP03Jn/k858BQuci7CLwjFIxf2muwIukhbScBGBHKdsm6jT9r/PeLnZOPZjPc+m7iN//EPCcRraBpAXvX/C0a0UCUhtTdeEtxgaJI9aeAwTr+5I0cPs2L6+BXFznXKdFoa1ufl8bvmlXDSSRYRTHyx/tL3YKV2+t/iBNUUonZasEryfVL/Z98Y7ckfaOSuxWen5FH9TZ8ui7WyKc8YIwRx49qxvjS59JS2YWk/3suvN2Ts8X346TUk3OzFKvihMM08JiwH/6C2oXhwxqdOspbEIA9RgmgPDKq4c4WCH3cASDPaTDFizXvpbqkMFPiKvMmBBchx02Rrwc6bRxBUGTvtQ/fnQbvZF8I8CrKgu7mQ+OiINZbFcYdY0Cjca+XUSnW6TAjKu8x7RZfmAikvZgagBYMoLQxYyrAuiNFAbR/jh1eiq09e8C00DOmBG9YzdSM8r7A1TgCYaeQZACqWysDVWkMLdxBRMQ26yxADAJujUFYkL6Pyr5534EI4ISIEBI8nUKTPMt3H3Jw8mO6I9Bw9//Q1quvXkM8EuIJhGVvlSKPXvzEO+MQiRK/OB9jijVZmz8yHndneImnjkm/5F9nQpO5sP6eYm88Q== jayam@JayamalDeVas"
}