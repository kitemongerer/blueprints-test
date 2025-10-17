ARG BASE_IMAGE
FROM debian:bookworm AS downloader
USER root
## See downloader.dockerfile.template
RUN apt-get update -y \
    && apt-get install -y --no-install-recommends \
        ca-certificates \
        curl \
        file \
        jq \
        unzip \
    && rm -rf /var/lib/apt/lists/*

FROM downloader AS doppler-downloader
ARG DOPPLER_VERSION="3.38.0"
USER root
## See doppler-downloader.dockerfile.template
ADD --link  https://github.com/DopplerHQ/cli/releases/download/${DOPPLER_VERSION}/doppler_${DOPPLER_VERSION}_linux_amd64.tar.gz /tmp/doppler.tar.gz
RUN echo '602bb2866dc7189de8f15d2bdeddff87d2aad0fffb5e1b61d43fd38c5c2d159e8a03fb37b623a662a502bc390a4c4376f171bb110b8377bcb69b338d709a3ce7 /tmp/doppler.tar.gz'| sha512sum -c - || exit 1
RUN mkdir /tmp/doppler
RUN tar -C /tmp/doppler -xf /tmp/doppler.tar.gz doppler
RUN chmod a+x /tmp/doppler/doppler

FROM downloader AS sentry-downloader
USER root
## See sentry-downloader.dockerfile.template
RUN mkdir /tmp/sentry
RUN curl --location -o /tmp/sentry/sentry-cli \
  $(curl --location https://api.github.com/repos/getsentry/sentry-cli/releases/latest | jq -r '.assets[] | select (.name | test("linux-x86_64"; "i")) | .browser_download_url')
RUN chmod a+x /tmp/sentry/sentry-cli

FROM downloader AS bun-downloader
ARG BUN_VERSION="1.2.20"
USER root
## See bun-downloader.dockerfile.template
ADD --link https://github.com/oven-sh/bun/releases/download/bun-v${BUN_VERSION}/bun-linux-x64-baseline.zip /tmp/bun.zip
RUN mkdir /tmp/bun
RUN unzip -j /tmp/bun.zip bun-linux-x64-baseline/bun -d /tmp/bun
RUN chmod a+x /tmp/bun/bun

FROM gcr.io/render-internal/docker-hub-mirror/buildpack-deps:bookworm

ARG NODE_VERSION=""
USER root
## See base.dockerfile.template
ENV LANG=C.UTF-8 \
    LC_ALL=C.UTF-8 \
    TERM=xterm-256color \
    DEBIAN_FRONTEND=noninteractive

RUN useradd --no-log-init --create-home --user-group --uid 1000 render \
  && usermod -d /opt/render render \
  && usermod --shell /bin/bash render

RUN apt-get -qq update \
  && apt-get -qq install --upgrade -y --no-install-recommends \
    apt-transport-https \
    apt-utils \
    bison \
    build-essential \
    ca-certificates \
    curl \
    dnsutils \
    ffmpeg \
    fonts-liberation \
    g++ \
    gcc \
    gconf-service \
    gettext \
    git \
    gnupg2 \
    jq \
    libappindicator1 \
    libasound2 \
    libatk-bridge2.0-0 \
    libatk1.0-0 \
    lsb-release \
    libc6 \
    libcairo2 \
    libcairo2-dev \
    libcups2 \
    libdbus-1-3 \
    libev-dev \
    libevdev2 \
    libexpat1 \
    libffi-dev \
    libfontconfig \
    libgcc1 \
    libgconf-2-4 \
    libgdk-pixbuf2.0-0 \
    libgif-dev \
    libglib2.0-0 \
    libgtk-3-0 \
    libjemalloc-dev \
    libjpeg-dev \
    libnspr4 \
    libnss3 \
    libpango-1.0-0 \
    libpangocairo-1.0-0 \
    libreadline-dev \
    libsodium-dev \
    libstdc++6 \
    libx11-6 \
    libx11-xcb1 \
    libxcb1 \
    libxcomposite1 \
    libxcursor1 \
    libxdamage1 \
    libxext6 \
    libxfixes3 \
    libxi6 \
    libxrandr2 \
    libxrender1 \
    libxss1 \
    libxtst6 \
    make \
    nano \
    pandoc \
    postgresql-client \
    python3-dev \
    python3-pip \
    python3-setuptools \
    rsync \
    software-properties-common \
    swig \
    sqlite3 \
    unzip \
    vim \
    wget \
    wkhtmltopdf \
    xdg-utils \
    xvfb \
    zip \
    zlib1g-dev \
  > /dev/null \
  && apt-get -qq clean \
  && rm -rf \
    /var/lib/apt/lists/* \
    /tmp/* \
    /var/tmp/* \
  && :

## See postgres.dockerfile.template
RUN echo "deb http://apt.postgresql.org/pub/repos/apt $(lsb_release -cs)-pgdg main" > /etc/apt/sources.list.d/pgdg.list \
  && wget --quiet -O - https://www.postgresql.org/media/keys/ACCC4CF8.asc | apt-key add - \
  && apt-get -qq update \
  && apt-get -qq install --upgrade -y --no-install-recommends \
    $(for v in $(seq 12 17); do echo postgresql-client-$v; done) \
  && apt-get -qq clean \
  && rm -rf \
    /var/lib/apt/lists/* \
    /tmp/* \
    /var/tmp/*

## See keyval.dockerfile.template
RUN echo "deb http://deb.debian.org/debian $(lsb_release -cs)-backports main" > /etc/apt/sources.list.d/redis.list \
&& apt-get -qq update \
&& apt-get -qq install --upgrade -y --no-install-recommends -t $(lsb_release -cs)-backports valkey-tools \
> /dev/null \
&& /bin/bash -c "for f in /usr/bin/valkey-*; do ln -s \$f \${f/valkey/redis}; done" \
&& apt-get -qq clean \
&& rm -rf \
  /var/lib/apt/lists/* \
  /tmp/* \
  /var/tmp/* \
  /etc/ImageMagick-*/policy.xml \
&& :

## See libvips.dockerfile.template
# lipvips installation as buster packages are outdated
# From https://github.com/libvips/libvips/wiki/Build-for-Ubuntu
#
RUN apt-get -qq update && apt-get -qq install -y --no-install-recommends \
    ninja-build \
    bc \
    libfftw3-dev \
    libopenexr-dev \
    libgsf-1-dev \
    libglib2.0-dev \
    liborc-dev \
    libopenslide-dev \
    libmatio-dev \
    libwebp-dev \
    libjpeg62-turbo-dev \
    libexpat1-dev \
    libexif-dev \
    libtiff5-dev \
    libcfitsio-dev \
    libpoppler-glib-dev \
    librsvg2-dev \
    libpango1.0-dev \
    libopenjp2-7-dev \
    liblcms2-dev \
    libimagequant-dev \
    autotools-dev \
    automake \
    libtool \
    meson

RUN wget https://github.com/libvips/libvips/releases/download/v8.15.3/vips-8.15.3.tar.xz \
  && tar xf vips-8.15.3.tar.xz \
  && cd vips-8.15.3 \
  && meson setup build --libdir=lib --buildtype=release -Dintrospection=disabled --prefix /usr \
  && cd build \
  && meson compile \
  && meson test \
  && meson install \
  && ldconfig \
  && cd ../.. \
  && rm -rf vips-8.15.3 vips-8.15.3.tar.xz

# End of libvips installation

## See node.dockerfile.template
RUN mkdir -p /etc/apt/keyrings && \
	curl -fsSL https://deb.nodesource.com/gpgkey/nodesource-repo.gpg.key | gpg --dearmor -o /etc/apt/keyrings/nodesource.gpg && \
	echo "deb [signed-by=/etc/apt/keyrings/nodesource.gpg] https://deb.nodesource.com/node_22.x nodistro main" | tee /etc/apt/sources.list.d/nodesource.list && \
	apt-get -qq update && apt-cache policy nodejs && apt-get -qq install -y nodejs=22.16.0-1nodesource1 > /dev/null && \
	npm install --global pnpm webpack yarn semver @renderinc/fetch-node-version typescript && \
	apt-get clean && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

ENV RENDER_NODE_INSTALLED=true
ENV NODE_VERSION=$NODE_VERSION

## See sharedrun.dockerfile.template
RUN apt-get -qq update \
  && apt-get -qq install --upgrade -y --no-install-recommends \
    gconf-service \
    gdebi \
    ghostscript \
    imagemagick \
    magic-wormhole \
    poppler-utils \
  > /dev/null \
  && apt-get -qq clean \
  && rm -rf \
    /var/lib/apt/lists/* \
    /tmp/* \
    /var/tmp/* \
  && :

## See princexml.dockerfile.template
# PrinceXML for Heroku users
RUN apt-get -qq update && \
  curl -o /tmp/princexml.deb \
  https://www.princexml.com/download/prince_15.3-1_debian12_amd64.deb && \
  gdebi --non-interactive /tmp/princexml.deb && \
  apt-get -qq clean && \
  rm -rf \
    /var/lib/apt/lists/* \
    /tmp/* \
    /var/tmp/*

USER 1000:1000
## See envwrapper.dockerfile.template
COPY --link --chown=1000:1000 envwrappers /home/render/envwrappers

RUN chmod a+x /home/render/envwrappers/*; \
  ln -s /home/render/envwrappers/wrapped-native-env /home/render/envwrappers/ruby; \
  ln -s /home/render/envwrappers/wrapped-native-env /home/render/envwrappers/gem; \
  ln -s /home/render/envwrappers/wrapped-native-env /home/render/envwrappers/bundle; \
  ln -s /home/render/envwrappers/wrapped-native-env /home/render/envwrappers/node; \
  ln -s /home/render/envwrappers/wrapped-native-env /home/render/envwrappers/yarn; \
  ln -s /home/render/envwrappers/wrapped-native-env /home/render/envwrappers/npm; \
  ln -s /home/render/envwrappers/wrapped-native-env /home/render/envwrappers/pnpm; \
  ln -s /home/render/envwrappers/wrapped-native-env /home/render/envwrappers/bun; \
  ln -s /home/render/envwrappers/wrapped-native-env /home/render/envwrappers/python; \
  ln -s /home/render/envwrappers/wrapped-native-env /home/render/envwrappers/pip; \
  ln -s /home/render/envwrappers/wrapped-native-env /home/render/envwrappers/pip3; \
  ln -s /home/render/envwrappers/wrapped-native-env /home/render/envwrappers/python3; \
  ln -s /home/render/envwrappers/wrapped-native-env /home/render/envwrappers/poetry; \
  ln -s /home/render/envwrappers/wrapped-native-env /home/render/envwrappers/uv

## See languageversions.dockerfile.template
# Store minimum and maintained language versions
RUN echo '{"erlang":"24.3.4","python-3":"3.7.3","ruby":"3.1.0"}' > /home/render/min-language-versions.json
RUN echo '{"erlang":"25.0.0","node":"20.0.0","python":"3.9.0","ruby":"3.2.0"}' > /home/render/maintained-language-versions.json

USER root
## See doppler.dockerfile.template
COPY --link --from=doppler-downloader /tmp/doppler/doppler   /usr/local/bin/doppler

## See sentry.dockerfile.template
COPY --link --from=sentry-downloader  /tmp/sentry/sentry-cli /usr/local/bin/sentry-cli

## See bun.dockerfile.template
COPY --link --chown=1000:1000 --from=bun-downloader     /tmp/bun/bun           /home/render/.bun/bin/bun
ENV PATH=/home/render/.bun/bin:$PATH

USER 1000:1000
