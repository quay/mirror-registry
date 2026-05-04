terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "6.18.1"
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

resource "google_compute_network" "vpc_network_extended_uninstall" {
  name = "terraform-network-extended-uninstall"
}

resource "google_compute_instance" "vm_instance_extended_uninstall" {
  name         = "mirror-ci-rhel-extended-uninstall"
  machine_type = "e2-standard-4"

  boot_disk {
    initialize_params {
      image = "rhel-9"
      size  = 100
    }
  }

  tags = ["mirror-ci-rhel-extended-uninstall"]

  network_interface {
    network = google_compute_network.vpc_network_extended_uninstall.name
    access_config {
    }
  }

  metadata = {
    ssh-keys = "ci-user:${var.SSH_PUBLIC_KEY}"
  }

  service_account {
    scopes = []
  }

  scheduling {
    max_run_duration {
      seconds = 7200
    }
    instance_termination_action = "DELETE"
  }
}

resource "google_compute_firewall" "ssh-rule-extended-uninstall" {
  name    = "vm-ssh-extended-uninstall"
  network = google_compute_network.vpc_network_extended_uninstall.name
  allow {
    protocol = "tcp"
    ports    = ["22", "80", "8080", "443", "8443"]
  }
  allow {
    protocol = "icmp"
  }
  target_tags   = ["mirror-ci-rhel-extended-uninstall"]
  source_ranges = ["0.0.0.0/0"]
}

output "ip" {
  value = google_compute_instance.vm_instance_extended_uninstall.network_interface.0.access_config.0.nat_ip
}
