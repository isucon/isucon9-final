use FindBin;
use lib "$FindBin::Bin/extlib/lib/perl5";
use lib "$FindBin::Bin/l";
use File::Basename;
use Plack::Builder;
use Kossy::Request;
use Isutrain::Web;

my $root_dir = File::Basename::dirname(__FILE__);

my $app = Isutrain::Web->psgi($root_dir);
builder {
    enable "Plack::Middleware::Log::Minimal", autodump => 1;
    enable 'ReverseProxy';
    enable 'Session::Cookie',
        session_key => 'session-isutrain',
        expires     => 3600,
        secret      => 'tagomoris';
    $app;
};
