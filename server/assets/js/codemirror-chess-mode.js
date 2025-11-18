// CodeMirror Chess PGN Mode
// Highlights chess notation with special styling for checks, checkmates, and current move

(function(mod) {
    if (typeof exports == "object" && typeof module == "object") // CommonJS
        mod(require("../../lib/codemirror"));
    else if (typeof define == "function" && define.amd) // AMD
        define(["../../lib/codemirror"], mod);
    else // Plain browser env
        mod(CodeMirror);
})(function(CodeMirror) {
    "use strict";

    CodeMirror.defineMode("chess", function() {
        return {
            token: function(stream, state) {
                // Skip whitespace
                if (stream.eatSpace()) return null;

                // Move numbers (e.g., "1.", "23.")
                if (stream.match(/^\d+\./)) {
                    return "chess-move-number";
                }

                // Ellipsis for black's move number (e.g., "1...")
                if (stream.match(/^\.\.\./)) {
                    return "chess-move-number";
                }

                // Checkmate (ends with #)
                if (stream.match(/^[KQRBN]?[a-h]?[1-8]?x?[a-h][1-8](=[QRBN])?#/)) {
                    return "chess-checkmate";
                }
                if (stream.match(/^O-O(-O)?#/)) {
                    return "chess-checkmate";
                }

                // Check (ends with +)
                if (stream.match(/^[KQRBN]?[a-h]?[1-8]?x?[a-h][1-8](=[QRBN])?\+/)) {
                    return "chess-check";
                }
                if (stream.match(/^O-O(-O)?\+/)) {
                    return "chess-check";
                }

                // Regular moves (piece moves, pawn moves, captures, castling)
                if (stream.match(/^[KQRBN][a-h]?[1-8]?x?[a-h][1-8]/)) {
                    return "chess-move";
                }
                if (stream.match(/^[a-h][1-8]/)) {
                    return "chess-move";
                }
                if (stream.match(/^[a-h]x[a-h][1-8]/)) {
                    return "chess-move";
                }
                if (stream.match(/^O-O(-O)?/)) {
                    return "chess-move";
                }

                // Promotion (e.g., "=Q")
                if (stream.match(/^=[QRBN]/)) {
                    return "chess-promotion";
                }

                // Game result
                if (stream.match(/^(1-0|0-1|1\/2-1\/2|\*)/)) {
                    return "chess-result";
                }

                // Annotations (!!, !, !?, ?!, ?, ??)
                if (stream.match(/^[!?]+/)) {
                    return "chess-annotation";
                }

                // Comments in braces
                if (stream.match(/^\{[^}]*\}/)) {
                    return "chess-comment";
                }

                // Variations in parentheses
                if (stream.match(/^[\(\)]/)) {
                    return "chess-variation";
                }

                // NAG (Numeric Annotation Glyph)
                if (stream.match(/^\$\d+/)) {
                    return "chess-nag";
                }

                // Skip any other character
                stream.next();
                return null;
            }
        };
    });

    CodeMirror.defineMIME("text/x-chess", "chess");
});
