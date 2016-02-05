/*
(**************************************************************************)
(*                                                                        *)
(*                                Schifra                                 *)
(*                Reed-Solomon Error Correcting Code Library              *)
(*                                                                        *)
(* Release Version 0.0.1                                                  *)
(* http://www.schifra.com                                                 *)
(* Copyright (c) 2000-2015 Arash Partow, All Rights Reserved.             *)
(*                                                                        *)
(* The Schifra Reed-Solomon error correcting code library and all its     *)
(* components are supplied under the terms of the General Schifra License *)
(* agreement. The contents of the Schifra Reed-Solomon error correcting   *)
(* code library and all its components may not be copied or disclosed     *)
(* except in accordance with the terms of that agreement.                 *)
(*                                                                        *)
(* URL: http://www.schifra.com/license.html                               *)
(*                                                                        *)
(**************************************************************************)
*/


/*
   Description: This example will demonstrate how to instantiate a Reed-Solomon
                encoder and decoder, add the full amount of possible errors,
                correct the errors, and output the various pieces of relevant
                information.
*/


// c++ -ansi -pedantic-errors -Wall -Wextra -Werror -Wno-long-long -O3 -o schifra_reed_solomon schifra_reed_solomon.cpp -lstdc++ -lm
// g++ -fPIC -O3 -c schifra_reed_solomon.cpp -lstdc++ -lm
// linux: g++ -shared -fPIC -o libschifra_reed_solomon.so schifra_reed_solomon.o
// mac: g++ -dynamiclib -fPIC -o libschifra_reed_solomon.dylib schifra_reed_solomon.o


#include <cstddef>
#include <iostream>
#include <string>

#include "schifra/schifra_galois_field.hpp"
#include "schifra/schifra_galois_field_polynomial.hpp"
#include "schifra/schifra_sequential_root_generator_polynomial_creator.hpp"
#include "schifra/schifra_reed_solomon_encoder.hpp"
#include "schifra/schifra_reed_solomon_decoder.hpp"
#include "schifra/schifra_reed_solomon_block.hpp"
#include "schifra/schifra_error_processes.hpp"

extern "C" void doRS(char *buf_in, char *buf_out);

void doRS(char *buf_in, char *buf_out) {
    /* Finite Field Parameters */
    const std::size_t field_descriptor                =   8;
    const std::size_t generator_polynomial_index      = 120;
    const std::size_t generator_polynomial_root_count =  20;

    /* Reed Solomon Code Parameters */
    const std::size_t code_length = 255;
    const std::size_t fec_length  =  20;
    const std::size_t data_length = code_length - fec_length;

   /* Instantiate Finite Field and Generator Polynomials */
   schifra::galois::field field(field_descriptor,
                                schifra::galois::primitive_polynomial_size06,
                                schifra::galois::primitive_polynomial06);

   schifra::galois::field_polynomial generator_polynomial(field);

   schifra::sequential_root_generator_polynomial_creator(field,
                                                         generator_polynomial_index,
                                                         generator_polynomial_root_count,
                                                         generator_polynomial);

   /* Instantiate Encoder (Codec) */
   schifra::reed_solomon::encoder<code_length,fec_length> encoder(field,generator_polynomial);

   //FIXME: Static length (temporary) for uplink messages.
   std::string message(buf_in, 72);
   // Pad with zeros.
   message = message + std::string(data_length - 72,static_cast<unsigned char>(0x00));


   /* Instantiate RS Block For Codec */
   schifra::reed_solomon::block<code_length,fec_length> block;

   /* Transform message into Reed-Solomon encoded codeword */
   if (!encoder.encode(message,block))
   {
      std::cout << "Error - Critical encoding failure!" << std::endl;
      return;
   }


   for (int i = 0; i < 255; i++) {
	   *(buf_out+i) = block[i];
   }

   return;
}
