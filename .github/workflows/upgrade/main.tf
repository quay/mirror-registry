terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "3.5.0"
    }
  }
}


variable "SSH_PUBLIC_KEY" {
  type = string
}

provider "google" {

  project = "quay-devel"
  region  = "us-central1"
  zone    = "us-central1-c"

}

resource "google_compute_network" "vpc_network_upgrade" {
  name = "terraform-network-upgrade"
}

resource "google_compute_instance" "vm_instance_upgrade" {
  name         = "mirror-ci-rhel-upgrade"
  machine_type = "e2-standard-8"

  boot_disk {
    initialize_params {
      image = "rhel-8"
    }
  }

  tags = ["mirror-ci-rhel-upgrade"]

  network_interface {
    network = google_compute_network.vpc_network_upgrade.name
    access_config {
    }
  }

  metadata = {
    ssh-keys = "jonathan:${var.SSH_PUBLIC_KEY}"
  }
}

resource "google_compute_firewall" "ssh-rule-upgrade" {
  name    = "vm-ssh-upgrade"
  network = google_compute_network.vpc_network_upgrade.name
  allow {
    protocol = "tcp"
    ports    = ["22", "80", "8080", "443", "8443"]
  }
  allow {
    protocol = "icmp"
  }
  target_tags   = ["mirror-ci-rhel-upgrade"]
  source_ranges = ["0.0.0.0/0"]
}

output "ip" {
  value = google_compute_instance.vm_instance_upgrade.network_interface.0.access_config.0.nat_ip
}
