package Isutrain::Parser::JSON;

use strict;
use warnings;
use JSON::MaybeXS qw/decode_json/;
use Encode qw/encode_utf8/;
use Data::Recursive::Encode;

sub new {
    bless [''], $_[0];
}

sub add {
    my $self = shift;
    if (defined $_[0]) {
        $self->[0] .= $_[0];
    }
}

sub finalize {
    my $self = shift;

    my $p = decode_json($self->[0]);
    $p = Data::Recursive::Encode->decode_utf8($p);
    my @params;
    if (ref $p eq 'HASH') {
        while (my ($k, $v) = each %$p) {
            push @params, $k, $v;
        }
    }
    return (\@params, []);
}

1;
