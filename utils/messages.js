const moment = require("moment");

function formatMessage(username, text, bot) {
  return {
    username,
    text,
    time: moment().format("h:mm a"),
    bot: bot
  };
}

module.exports = formatMessage;
