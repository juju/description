# Description

Describes the Juju 2.0 serialization format of a model

-----

The description package is a representation of a Juju model. Over the wire
format of the Juju model is intended to be yaml.

The design of the description package from the outset supports independent
versioning of entities. Each entity can therefor change without rev'ing other
entities modelled with in the serialized format.

In this contrived example, it's possible to bump the status entity without
bumping the applications. If how ever the entity in question requires a change
with application or other entities, those also will need to be bumped.

```yaml
applications:
  applications:
  - name: ubuntu
    status:
      status:
        message: waiting for machine
      version: 1
  version: 1
```

-----

The concept of description package in the purest sense, is to ensure that it's
possible to encode and decode any entity for the right version. How each version
is then correctly implemented is out of scope of the description package.