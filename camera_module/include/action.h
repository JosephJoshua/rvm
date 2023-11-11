#pragma once

#include <WString.h>
#include <variant>

namespace action
{
    struct None { };

    struct ClassifyImage { };

    struct PublishImageToMQTTBroker
    {
        String request_id;
        int tries;
    };
}

using Action = std::variant<
    action::None,
    action::ClassifyImage,
    action::PublishImageToMQTTBroker
>;

template<class... Ts>
struct overloaded : Ts... { using Ts::operator()...; };

// Some compilers might require this explicit deduction guide
template<class... Ts>
overloaded(Ts...) -> overloaded<Ts...>;
