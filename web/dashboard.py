# Import packages
import streamlit as st
import pandas as pd
import os
import hmac


def load_commute_time(file):
    try:
        df = pd.read_csv(file, names=["datetime", "commute_time"])
        df_dt = pd.to_datetime(df["datetime"])
        df["hour_min"] = df_dt.dt.strftime("%H:%M")
        df_weekday_nbr = df_dt.dt.weekday + 1
        df_weekday_name = df_dt.dt.day_name()
        df["weekday"] = df_weekday_nbr.astype(str) + " - " + df_weekday_name
    except Exception as e:
        print(e)
        df = pd.DataFrame({"hour_min": [], "commute_time": []})
    return df


def get_average_commute_time(df: pd.DataFrame):
    avg_time = df.groupby("hour_min")["commute_time"].mean()
    avg_time = avg_time.reset_index()
    return avg_time


def get_average_commute_time_daywise(df: pd.DataFrame):
    avg_time_daywise = df.groupby(["weekday", "hour_min"])["commute_time"].mean()
    avg_time_daywise = avg_time_daywise.reset_index()
    return avg_time_daywise


def check_password():
    """Returns `True` if the user had the correct password."""

    def password_entered():
        """Checks whether a password entered by the user is correct."""
        if hmac.compare_digest(
            st.session_state["password"], os.environ["DASHBOARD_PASSWORD"]
        ):
            st.session_state["password_correct"] = True
            del st.session_state["password"]  # Don't store the password.
        else:
            st.session_state["password_correct"] = False

    # Return True if the password is validated.
    if st.session_state.get("password_correct", False):
        return True

    # Show input for password.
    st.text_input(
        "Password", type="password", on_change=password_entered, key="password"
    )
    if "password_correct" in st.session_state:
        st.error("😕 Password incorrect")
    return False


# Run the app
if __name__ == "__main__":
    st.set_page_config(page_title="Commute time dashboard", layout="wide")

    if "DASHBOARD_PASSWORD" in os.environ and not check_password():
        st.stop()  # Do not continue if check_password is not True.

    col1, col2 = st.columns(2, border=True)
    with col1:
        st.header("Commute time - to")
        df = load_commute_time("data/to.csv")

        st.subheader("Average commute time")
        adf = get_average_commute_time(df)
        adf["+/-"] = df.groupby("hour_min")["commute_time"].std().values
        st.bar_chart(
            adf,
            x="hour_min",
            y="commute_time",
            x_label="Departure time (HH:MM)",
            y_label="Average commute time (min)",
            color="+/-",
        )

        st.subheader("Average day-wise commute time")
        adf = get_average_commute_time_daywise(df)
        st.bar_chart(
            adf,
            x="hour_min",
            y="commute_time",
            x_label="Departure time (HH:MM)",
            y_label="Average commute time (min)",
            stack=False,
            color="weekday",
        )
    if st.checkbox("Show 'to' data"):
        st.dataframe(df)

    with col2:
        st.header("Commute time - from")
        df = load_commute_time("data/from.csv")

        st.subheader("Average commute time")
        adf = get_average_commute_time(df)
        adf["+/-"] = df.groupby("hour_min")["commute_time"].std().values
        st.bar_chart(
            adf,
            x="hour_min",
            y="commute_time",
            x_label="Departure time (HH:MM)",
            y_label="Average commute time (min)",
            color="+/-",
        )

        st.subheader("Average day-wise commute time")
        adf = get_average_commute_time_daywise(df)
        st.bar_chart(
            adf,
            x="hour_min",
            y="commute_time",
            x_label="Departure time (HH:MM)",
            y_label="Average commute time (min)",
            stack=False,
            color="weekday",
        )

    if st.checkbox("Show 'from' data"):
        st.dataframe(df)
