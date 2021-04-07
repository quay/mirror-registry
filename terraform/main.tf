terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "3.5.0"
    }
  }
}

provider "google" {

  credentials = file("terraform-key.json")

  project = "quay-devel"
  region  = "us-central1"
  zone    = "us-central1-c"
}

resource "google_compute_network" "vpc_network" {
  name = "terraform-network"
}

resource "google_compute_instance" "vm_instance" {
  name         = "jonathan-rhel-1"
  machine_type = "e2-medium"

  boot_disk {
    initialize_params {
      image = "rhel-8"
    }
  }

  metadata = {
    ssh-keys = "jonathan:${file("~/.ssh/gcp.pub")}"
  }

  tags = ["jonathan-rhel-1"]

  network_interface {
    network = google_compute_network.vpc_network.name
    access_config {
    }
  }
}

resource "google_compute_firewall" "ssh-rule" {
  name    = "vm-ssh"
  network = google_compute_network.vpc_network.name
  allow {
    protocol = "tcp"
    ports    = ["22", "80", "8080", "443", "8443"]
  }
  allow {
    protocol = "icmp"
  }
  target_tags   = ["jonathan-rhel-1"]
  source_ranges = ["0.0.0.0/0"]
}
