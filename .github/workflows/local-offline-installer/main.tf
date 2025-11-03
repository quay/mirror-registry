terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 6.0"
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

resource "google_compute_network" "vpc_network_local_offline_install" {
  name = "terraform-network-local-offline-install"
}

resource "google_compute_instance" "vm_instance_local_offline_install" {
  name         = "mirror-ci-rhel-local-offline-install"
  machine_type = "e2-standard-16"

  boot_disk {
    initialize_params {
      image = "rhel-9"
      size  = 100
    }
  }

  tags = ["mirror-ci-rhel-local-offline-install"]

  network_interface {
    network = google_compute_network.vpc_network_local_offline_install.name
    access_config {
    }
  }

  metadata = {
    ssh-keys = "jonathan:${var.SSH_PUBLIC_KEY}"
  }
}

resource "google_compute_firewall" "ssh-rule-local-offline-install" {
  name    = "vm-ssh-local-offline-install"
  network = google_compute_network.vpc_network_local_offline_install.name
  allow {
    protocol = "tcp"
    ports    = ["22", "80", "8080", "443", "8443"]
  }
  allow {
    protocol = "icmp"
  }
  target_tags   = ["mirror-ci-rhel-local-offline-install"]
  source_ranges = ["0.0.0.0/0"]
}

output "ip" {
  value = google_compute_instance.vm_instance_local_offline_install.network_interface.0.access_config.0.nat_ip
}
