data "terraform_remote_state" "dest" {
  backend = "local"

  config = {
    path = "../destinations/terraform.tfstate"
  }
}

resource "lightstep_alert" "alert" { 
    name = "Alert"
    project_name = "dev-heidmo"
    
    expression {
        is_multi   = false
        is_no_data = false
        operand    = "above"
        thresholds {
            warning  = 5.0
            critical = 10.0
        }
    }

    query {
        query_name = "a"
        hidden = true
        query_string = <<-EOF
        metric foo
        | rate 15s
        EOF
    }

    alerting_rule {
      id = data.terraform_remote_state.dest.outputs.hook
    }
}