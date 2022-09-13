# Frequently Asked Questions (FAQs)

> Why should I use Koko instead of the Kong Gateway's built-in control plane?

Koko is the second generation Control-Plane (CP) for Kong Gateway. All CP
aspects of Kong have been redesigned to offer a better experience and make it
easier to operate the CP. Some highlights of differences between Koko and built-in Kong Control Plane:
- Simplified integration and programmatic access
- Freedom to choose the database of your choice
- Easy migrations between Kong Gateway versions and huge focus on
  backwards compatibility across major versions of Kong Gateway
- Improved operational model with structured logs, metrics and planned OpenTracing support

For a complete list, please read [comparison document](./koko-vs-kong.md).

> Why was Koko developed & what problems is it trying to solve?

Kong Gateway project started in 2014. Infrastructure software space has seen
major shifts and changes since then. Abstractions and assumptions baked into
Kong have held the test of time but some of them are showing age.
In addition to that, Kong's underlying technology stack of OpenResty and NGINX
is a good choice for a Data-Plane proxy where performance is paramount. The same
is not the case for Control-Plane aspects.
The ecosystem of Lua is not focused on building Control-Plane software, which
increases the cost to build features and furthermore increases maintenance overhead.

> What Kong Gateway versions are supported?

Kong Gateway versions 2.5 and above are supported.
The recommended version for use is 3.0 or above.

> What is the expected release cadence?

There is no set release cadence for Koko yet. You can expect one minor release
every couple of months which will be feature heavy and patch releases throughout
the year as adoption increases.

> Does Koko support all features, including plugins, of Kong Gateway?

_Most_ features are supported with a select number of supported plugins.
Please refer to the
[Comparison of Kong Gateway Control-Plane with Koko](./koko-vs-kong.md) document.

> How can I migrate from an existing Kong Gateway deployment to Koko?

Work to support migrations is underway. This will be accomplished by two features:
- A backwards compatible Admin API
- Support for [decK](https://github.com/kong/deck) in Koko

Kong Gateway setups can be customized in a number of ways so every detail
matters. Additional documentation around migration is planned.

> I'm happy with Kong's traditional mode. Why should I use Koko?

If you have set up Kong and are happy with your setup, you can continue to use
Kong the way you have been using so far.
However, if you are interested in using one of the features of Koko,
you will need to migrate.

## Further help

If your question was not answered or the answer provided here is not sufficient,
please open a [GitHub Issue](https://github.com/Kong/koko/issues/new/choose).

