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

resource "google_compute_network" "vpc_network_remote_online_install" {
  name = "terraform-network-remote-online-install"
}

resource "google_compute_instance" "control_vm_remote_online_install" {
  name         = "mirror-ci-control-remote-online-install"
  machine_type = "e2-standard-16"

  boot_disk {
    initialize_params {
      image = "rhel-9"
      size  = 100
    }
  }

  tags = ["mirror-ci-rhel-remote-online-install"]

  network_interface {
    network = google_compute_network.vpc_network_remote_online_install.name
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

resource "google_compute_instance" "target_vm_remote_online_install" {
  name         = "mirror-ci-target-remote-online-install"
  machine_type = "e2-standard-16"

  boot_disk {
    initialize_params {
      image = "rhel-9"
      size  = 100
    }
  }

  tags = ["mirror-ci-rhel-remote-online-install"]

  network_interface {
    network = google_compute_network.vpc_network_remote_online_install.name
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

resource "google_compute_firewall" "ssh-rule-remote-online-install" {
  name    = "vm-ssh-remote-online-install"
  network = google_compute_network.vpc_network_remote_online_install.name
  allow {
    protocol = "tcp"
    ports    = ["22", "80", "8080", "443", "8443"]
  }
  allow {
    protocol = "icmp"
  }
  target_tags   = ["mirror-ci-rhel-remote-online-install"]
  source_ranges = ["0.0.0.0/0"]
}

output "ip" {
  value = google_compute_instance.target_vm_remote_online_install.network_interface.0.access_config.0.nat_ip
}

output "control_ip" {
  value = google_compute_instance.control_vm_remote_online_install.network_interface.0.access_config.0.nat_ip
}
