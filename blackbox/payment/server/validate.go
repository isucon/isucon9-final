package server

import (
	pb "payment/pb"
	"time"
	"strconv"
	"strings"
	"errors"
	"regexp"
)

func (s *Server) ValidateCardInformation(req *pb.RegistCardRequest) error {
	card := pb.CardInformation{
		CardNumber: req.CardInformation.CardNumber,
		Cvv:        req.CardInformation.Cvv,
		ExpiryDate: req.CardInformation.ExpiryDate,
	}

	if len(card.CardNumber) != 8 {
		return errors.New("Invalid CardNumber Length")
	}
	if len(card.Cvv) != 3 {
		return errors.New("Invalid Cvv Length")
	}
	if len(card.ExpiryDate) != 5 {
		return errors.New("Invalid ExpiryDate length")
	}

	cardnum := regexp.MustCompile("^[0-9]{8}$")
	if !cardnum.MatchString(card.CardNumber) {
		return errors.New("Invalid CardNumber")
	}
	cvvnum := regexp.MustCompile("^[0-9]{3}$")
	if !cvvnum.MatchString(card.Cvv) {
		return errors.New("Invalid Cvv")
	}
	expdate := regexp.MustCompile("^[0-9]{2}/[0-9]{2}$")
	if !expdate.MatchString(card.ExpiryDate) {
		return errors.New("Invalid ExpiryDate")
	}

	mmyy := strings.Split(card.ExpiryDate, "/")
	month, err := strconv.Atoi(mmyy[0])
	if err != nil {
		return err
	}
	y := strconv.Itoa(time.Now().UTC().Year())
	if err != nil {
		return err
	}
	year, err := strconv.Atoi(y[:2] + mmyy[1])
	if err != nil {
		return err
	}

	if month < 1 || 12 < month {
		return errors.New("Invalid month.")
	}
	if year < time.Now().UTC().Year() {
		return errors.New("Credit card has expired.")
	}
	if year == time.Now().UTC().Year() && month < int(time.Now().UTC().Month()) {
		return errors.New("Credit card has expired.")
	}
	return nil
}
