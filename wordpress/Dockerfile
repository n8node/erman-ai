FROM wordpress:6-php8.3-apache

RUN curl -O https://raw.githubusercontent.com/wp-cli/builds/gh-pages/phar/wp-cli.phar && \
    chmod +x wp-cli.phar && \
    mv wp-cli.phar /usr/local/bin/wp

COPY conf/forwarded-https.conf /etc/apache2/conf-available/forwarded-https.conf
RUN a2enconf forwarded-https
