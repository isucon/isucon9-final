use FindBin;
use lib "$FindBin::Bin/extlib/lib/perl5";
use lib "$FindBin::Bin/l";
use File::Basename;
use Plack::Builder;
use Kossy::Request;
use Isutrain::Web;

my $default_parser = HTTP::Entity::Parser->new();
$default_parser->register(
    'application/x-www-form-urlencoded',
    'HTTP::Entity::Parser::UrlEncoded'
);
$default_parser->register(
    'multipart/form-data',
    'HTTP::Entity::Parser::MultiPart'
);

my $json_parser = HTTP::Entity::Parser->new();
$json_parser->register(
    'application/x-www-form-urlencoded',
    'HTTP::Entity::Parser::UrlEncoded'
);
$json_parser->register(
    'multipart/form-data',
    'HTTP::Entity::Parser::MultiPart'
);
$json_parser->register(
    'application/json',
    'Isutrain::Parser::JSON'
);

sub Kossy::Request::_build_request_body_parser {
    my $self = shift;
    if ( $self->env->{'kossy.request.parse_json_body'} ) {
        return $json_parser;
    }
    $default_parser;
}

my $root_dir = File::Basename::dirname(__FILE__);

my $app = Isutrain::Web->psgi($root_dir);
builder {
    enable 'ReverseProxy';
    enable 'Session::Cookie',
        session_key => 'session-isutrain',
        expires     => 3600,
        secret      => 'tagomoris';
    $app;
};
