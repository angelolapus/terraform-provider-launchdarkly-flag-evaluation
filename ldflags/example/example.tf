terraform {
  required_providers {
    ldflags = {
      version = "0.2"
      source  = "launchdarkly/ldflags"
    }
  }
}

provider "ldflags" {
  sdk_key = "{your_sdk_key}"
}

data "ldflags_evaluation_boolean" "mybool" {
  flag_key      = "boolean-flag"
  default_value = true
  context = {
    key = "hosh1@launchdarkly.com"
    custom = {
      "test"  = "foo"
      "test2" = 123
      "test3" = false
      "test4" = ["red", "orange", "yellow", true, 4] // tuple example
      "test5" = tolist(["b", "w"])                   // list example
      // note: cannot pass JSON objects as custom properties
      // https://docs.launchdarkly.com/home/users/attributes
    }
  }
}

data "ldflags_evaluation_string" "mystring" {
  flag_key      = "string-flag"
  default_value = "def"
  context = {
    key = "mchheda@launchdarkly.com"
    custom = {
      "test"  = "bar"
      "test2" = 456
      "test3" = false
      "test4" = ["black", "white", true, 4] // tuple example
      "test5" = tolist(["a", "d"])          // list example
      "test6" = 100.50
      // note: cannot pass JSON objects as custom properties
      // https://docs.launchdarkly.com/home/users/attributes
    }
  }
}

data "ldflags_evaluation_int" "myint" {
  flag_key      = "int-flag"
  default_value = 1
  context = {
    key = "mwong@launchdarkly.com"
    custom = {
      "test"  = "bar"
      "test2" = 789
      "test3" = false
      "test4" = ["red", "green", true, 4] // tuple example
      "test5" = tolist(["c", "f"])        // list example
      "test6" = 2.2
      // note: cannot pass JSON objects as custom properties
      // https://docs.launchdarkly.com/home/users/attributes
    }
  }
}

data "ldflags_evaluation_float" "myfloat" {
  flag_key      = "float-64-flag"
  default_value = 1
  context = {
    key = "mwong@launchdarkly.com"
    custom = {
      "test"  = "bar"
      "test2" = 789
      "test3" = false
      "test4" = ["red", "green", true, 4] // tuple example
      "test5" = tolist(["c", "f"])        // list example
      "test6" = 6.5
      // note: cannot pass JSON objects as custom properties
      // https://docs.launchdarkly.com/home/users/attributes
    }
  }
}

data "ldflags_evaluation_json" "myjson" {
  flag_key      = "json-flag"
  default_value = jsonencode({
    "id": "moonshot",
    "value": "tf-dont-eval"
  })
  context = {
    key = "mmchheda@launchdarkly.com"
    custom = {
      "test"  = "foo"
      "test2" = 789
      "test3" = false
      "test4" = ["red", "orange", "black", true, 4] // tuple example
      "test5" = tolist(["f", "w"])                   // list example
      "test6" = 6.5
      // note: cannot pass JSON objects as custom properties
      // https://docs.launchdarkly.com/home/users/attributes
    }
  }
}


output "variation_value" {
  value = data.ldflags_evaluation_boolean.mybool.value
}

output "variation_value_string" {
  value = data.ldflags_evaluation_string.mystring.value
}

output "variation_value_int" {
  value = data.ldflags_evaluation_int.myint.value
}

output "variation_value_float" {
  value = data.ldflags_evaluation_float.myfloat.value
}

output "variation_value_json" {
  value = jsondecode(data.ldflags_evaluation_json.myjson.value)
}


# locals {
#   foo = file(data.ldflags_boolean.mybool.flag_key)
# }
