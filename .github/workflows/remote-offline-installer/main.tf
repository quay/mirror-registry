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

resource "google_compute_network" "vpc_network_remote_offline_install" {
  name = "terraform-network-remote-offline-install"
}

resource "google_compute_instance" "vm_instance_remote_offline_install" {
  name         = "mirror-ci-rhel-remote-offline-install"
  machine_type = "e2-standard-16"

  boot_disk {
    initialize_params {
      image = "rhel-8"
      size  = 100
    }
  }

  tags = ["mirror-ci-rhel-remote-offline-install"]

  network_interface {
    network = google_compute_network.vpc_network_remote_offline_install.name
    access_config {
    }
  }

  metadata = {
    ssh-keys = "jonathan:${var.SSH_PUBLIC_KEY}"
  }
}

resource "google_compute_firewall" "ssh-rule-remote-offline-install" {
  name    = "vm-ssh-remote-offline-install"
  network = google_compute_network.vpc_network_remote_offline_install.name
  allow {
    protocol = "tcp"
    ports    = ["22", "80", "8080", "443", "8443"]
  }
  allow {
    protocol = "icmp"
  }
  target_tags   = ["mirror-ci-rhel-remote-offline-install"]
  source_ranges = ["0.0.0.0/0"]
}

output "ip" {
  value = google_compute_instance.vm_instance_remote_offline_install.network_interface.0.access_config.0.nat_ip
}
